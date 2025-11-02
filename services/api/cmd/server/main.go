package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
	authmiddleware "github.com/sergeirastrigin/ubik-enterprise/services/api/internal/middleware"
)

func main() {
	// Get configuration from environment
	dbURL := getEnv("DATABASE_URL", "postgres://ubik:ubik_dev_password@localhost:5432/ubik?sslmode=disable")
	port := getEnv("PORT", "3001")

	// Connect to database
	ctx := context.Background()
	log.Printf("Connecting to database: %s", maskPassword(dbURL))

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// Create database queries instance
	queries := db.New(dbPool)

	// Create handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(queries)
	employeesHandler := handlers.NewEmployeesHandler(queries)
	rolesHandler := handlers.NewRolesHandler(queries)
	organizationsHandler := handlers.NewOrganizationsHandler(queries)
	teamsHandler := handlers.NewTeamsHandler(queries)
	agentsHandler := handlers.NewAgentsHandler(queries)
	orgAgentConfigsHandler := handlers.NewOrgAgentConfigsHandler(queries)
	teamAgentConfigsHandler := handlers.NewTeamAgentConfigsHandler(queries)
	employeeAgentConfigsHandler := handlers.NewEmployeeAgentConfigsHandler(queries)
	activityLogsHandler := handlers.NewActivityLogsHandler(queries)
	subscriptionsHandler := handlers.NewSubscriptionsHandler(queries)
	usageStatsHandler := handlers.NewUsageStatsHandler(queries)
	agentRequestsHandler := handlers.NewAgentRequestsHandler(queries)
	claudeTokensHandler := handlers.NewClaudeTokensHandler(queries)
	mcpServersHandler := handlers.NewMCPServersHandler(queries)
	// skillsHandler := handlers.NewSkillsHandler(queries) // TODO: Re-enable when Skills API is complete (PR #66)

	// Setup router
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Helper function to serve static files with no-cache in dev mode
	serveStaticFile := func(filepath string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Disable caching for development
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			http.ServeFile(w, r, filepath)
		}
	}

	// Serve static files (HTML prototype)
	fileServer := http.FileServer(http.Dir("./static"))
	router.Handle("/static/*", http.StripPrefix("/static", fileServer))
	router.Get("/", serveStaticFile("./static/login.html"))
	router.Get("/components.html", serveStaticFile("./static/components.html"))
	router.Get("/login.html", serveStaticFile("./static/login.html"))
	router.Get("/dashboard.html", serveStaticFile("./static/dashboard.html"))
	router.Get("/employees.html", serveStaticFile("./static/employees.html"))
	router.Get("/teams.html", serveStaticFile("./static/teams.html"))
	router.Get("/agents.html", serveStaticFile("./static/agents.html"))
	router.Get("/settings.html", serveStaticFile("./static/settings.html"))
	router.Get("/profile.html", serveStaticFile("./static/profile.html"))
	router.Get("/employee-detail.html", serveStaticFile("./static/employee-detail.html"))
	router.Get("/team-detail.html", serveStaticFile("./static/team-detail.html"))
	router.Get("/create-employee.html", serveStaticFile("./static/create-employee.html"))
	router.Get("/roles.html", serveStaticFile("./static/roles.html"))
	router.Get("/employee-agent-configs.html", serveStaticFile("./static/employee-agent-configs.html"))
	router.Get("/add-employee-agent-config.html", serveStaticFile("./static/add-employee-agent-config.html"))
	router.Get("/edit-employee-agent-config.html", serveStaticFile("./static/edit-employee-agent-config.html"))

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		// Health check
		r.Get("/health", healthHandler.HealthCheck)

		// Auth routes (login is public, others need auth)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(authmiddleware.JWTAuth(queries))
				r.Post("/logout", authHandler.Logout)
				r.Get("/me", authHandler.GetMe)
			})
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.JWTAuth(queries))

			// Employees routes
			r.Route("/employees", func(r chi.Router) {
				r.Get("/", employeesHandler.ListEmployees)
				r.Post("/", employeesHandler.CreateEmployee)
				r.Get("/{employee_id}", employeesHandler.GetEmployee)
				r.Patch("/{employee_id}", employeesHandler.UpdateEmployee)
				r.Delete("/{employee_id}", employeesHandler.DeleteEmployee)

				// Employee personal Claude token (hybrid auth)
				r.Put("/me/claude-token", claudeTokensHandler.SetEmployeeClaudeToken)
				r.Delete("/me/claude-token", claudeTokensHandler.DeleteEmployeeClaudeToken)
				r.Get("/me/claude-token/status", claudeTokensHandler.GetClaudeTokenStatus)
				r.Get("/me/claude-token/effective", claudeTokensHandler.GetEffectiveClaudeToken)
			})

			// Roles routes
			r.Route("/roles", func(r chi.Router) {
				r.Get("/", rolesHandler.ListRoles)
				r.Post("/", rolesHandler.CreateRole)
				r.Get("/{role_id}", rolesHandler.GetRole)
				r.Patch("/{role_id}", rolesHandler.UpdateRole)
				r.Delete("/{role_id}", rolesHandler.DeleteRole)
			})

			// Teams routes
			r.Route("/teams", func(r chi.Router) {
				r.Get("/", teamsHandler.ListTeams)
				r.Post("/", teamsHandler.CreateTeam)
				r.Get("/{team_id}", teamsHandler.GetTeam)
				r.Patch("/{team_id}", teamsHandler.UpdateTeam)
				r.Delete("/{team_id}", teamsHandler.DeleteTeam)

				// Team agent configs
				r.Route("/{team_id}/agent-configs", func(r chi.Router) {
					r.Get("/", teamAgentConfigsHandler.ListTeamAgentConfigs)
					r.Post("/", teamAgentConfigsHandler.CreateTeamAgentConfig)
					r.Get("/{config_id}", teamAgentConfigsHandler.GetTeamAgentConfig)
					r.Patch("/{config_id}", teamAgentConfigsHandler.UpdateTeamAgentConfig)
					r.Delete("/{config_id}", teamAgentConfigsHandler.DeleteTeamAgentConfig)
				})
			})

			// Organizations routes
			r.Route("/organizations/current", func(r chi.Router) {
				r.Get("/", organizationsHandler.GetCurrentOrganization)
				r.Patch("/", organizationsHandler.UpdateCurrentOrganization)

				// Organization Claude token (hybrid auth)
				r.Put("/claude-token", claudeTokensHandler.SetOrganizationClaudeToken)
				r.Delete("/claude-token", claudeTokensHandler.DeleteOrganizationClaudeToken)

				// Organization agent configs
				r.Route("/agent-configs", func(r chi.Router) {
					r.Get("/", orgAgentConfigsHandler.ListOrgAgentConfigs)
					r.Post("/", orgAgentConfigsHandler.CreateOrgAgentConfig)
					r.Get("/{config_id}", orgAgentConfigsHandler.GetOrgAgentConfig)
					r.Patch("/{config_id}", orgAgentConfigsHandler.UpdateOrgAgentConfig)
					r.Delete("/{config_id}", orgAgentConfigsHandler.DeleteOrgAgentConfig)
				})
			})

			// Employee agent configs routes
			r.Route("/employees/{employee_id}/agent-configs", func(r chi.Router) {
				r.Get("/", employeeAgentConfigsHandler.ListEmployeeAgentConfigs)
				r.Post("/", employeeAgentConfigsHandler.CreateEmployeeAgentConfig)
				r.Get("/resolved", orgAgentConfigsHandler.GetEmployeeResolvedAgentConfigs)
				r.Get("/{config_id}", employeeAgentConfigsHandler.GetEmployeeAgentConfig)
				r.Patch("/{config_id}", employeeAgentConfigsHandler.UpdateEmployeeAgentConfig)
				r.Delete("/{config_id}", employeeAgentConfigsHandler.DeleteEmployeeAgentConfig)
			})

			// Agents catalog routes (read-only)
			r.Route("/agents", func(r chi.Router) {
				r.Get("/", agentsHandler.ListAgents)
				r.Get("/{agent_id}", agentsHandler.GetAgent)
			})

			// MCP servers catalog routes
			r.Route("/mcp-servers", func(r chi.Router) {
				r.Get("/", mcpServersHandler.ListMCPServers)
				r.Get("/{id}", mcpServersHandler.GetMCPServer)
			})

			// Employee MCP servers routes
			r.Route("/employees/me/mcp-servers", func(r chi.Router) {
				r.Get("/", mcpServersHandler.ListEmployeeMCPServers)
			})

			// TODO: Re-enable when Skills API is complete (PR #66)
			// Skills catalog routes
			// r.Route("/skills", func(r chi.Router) {
			// 	r.Get("/", skillsHandler.ListSkills)
			// 	r.Get("/{skill_id}", skillsHandler.GetSkill)
			// })

			// Employee skills routes
			// r.Route("/employees/me/skills", func(r chi.Router) {
			// 	r.Get("/", skillsHandler.ListEmployeeSkills)
			// 	r.Get("/{skill_id}", skillsHandler.GetEmployeeSkill)
			// })

			// Activity logs routes
			r.Route("/activity-logs", func(r chi.Router) {
				r.Get("/", activityLogsHandler.ListActivityLogs)
			})

			// Subscription routes
			r.Route("/organizations/current/subscription", func(r chi.Router) {
				r.Get("/", subscriptionsHandler.GetCurrentSubscription)
			})

			// Usage stats routes
			r.Route("/usage-stats", func(r chi.Router) {
				r.Get("/org", usageStatsHandler.GetOrgUsageStats)
				r.Get("/me", usageStatsHandler.GetCurrentEmployeeUsageStats)
			})

			r.Route("/employees/{employee_id}/usage-stats", func(r chi.Router) {
				r.Get("/", usageStatsHandler.GetEmployeeUsageStats)
			})

			// Agent requests routes
			r.Route("/agent-requests", func(r chi.Router) {
				r.Get("/", agentRequestsHandler.ListAgentRequests)
				r.Get("/pending/count", agentRequestsHandler.GetPendingCount)
			})
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s", port)
		log.Printf("🌐 Web UI: http://localhost:%s/", port)
		log.Printf("📝 API Documentation: http://localhost:%s/api/v1/health", port)
		log.Printf("🔐 Auth endpoints:")
		log.Printf("   POST http://localhost:%s/api/v1/auth/login", port)
		log.Printf("   POST http://localhost:%s/api/v1/auth/logout", port)
		log.Printf("   GET  http://localhost:%s/api/v1/auth/me", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// maskPassword masks the password in a database URL for logging
func maskPassword(dbURL string) string {
	// Simple masking - just show first 20 chars
	if len(dbURL) > 40 {
		return dbURL[:20] + "***" + dbURL[len(dbURL)-10:]
	}
	return "***"
}
