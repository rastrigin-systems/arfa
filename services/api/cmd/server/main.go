package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/rastrigin-systems/ubik-enterprise/generated/api"
	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/handlers"
	authmiddleware "github.com/rastrigin-systems/ubik-enterprise/services/api/internal/middleware"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/service"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/websocket"
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

	// Create WebSocket hub for real-time log streaming
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Create handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(queries)
	employeesHandler := handlers.NewEmployeesHandler(queries)
	rolesHandler := handlers.NewRolesHandler(queries)
	organizationsHandler := handlers.NewOrganizationsHandler(queries)
	teamsHandler := handlers.NewTeamsHandler(queries)
	activityLogsHandler := handlers.NewActivityLogsHandler(queries)
	logsHandler := handlers.NewLogsHandler(queries, wsHub)
	subscriptionsHandler := handlers.NewSubscriptionsHandler(queries)
	wsHandler := websocket.NewHandler(wsHub)
	usageStatsHandler := handlers.NewUsageStatsHandler(queries)
	agentRequestsHandler := handlers.NewAgentRequestsHandler(queries)
	claudeTokensHandler := handlers.NewClaudeTokensHandler(queries)
	skillsHandler := handlers.NewSkillsHandler(queries)
	toolPoliciesHandler := handlers.NewToolPoliciesHandler(queries)
	webhooksHandler := handlers.NewWebhooksHandler(queries)

	// Email service (MockEmailService for development)
	emailService := service.NewMockEmailService()
	invitationsHandler := handlers.NewInvitationHandler(queries, emailService)

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
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"https://ubik-web-754414213269.us-central1.run.app",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API Documentation (public, no auth required)
	router.Handle("/api/docs/*", handlers.SwaggerHandler())
	router.Get("/api/docs/spec.yaml", handlers.SpecHandler())

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		// Health check
		r.Get("/health", healthHandler.HealthCheck)

		// Auth routes (login and register are public, others need auth)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)
			r.Get("/check-slug", authHandler.CheckSlugAvailability)

			// Password reset routes (public)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Get("/verify-reset-token", authHandler.VerifyResetToken)
			r.Post("/reset-password", authHandler.ResetPassword)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(authmiddleware.JWTAuth(queries))
				r.Post("/logout", authHandler.Logout)
				r.Get("/me", authHandler.GetMe)
			})
		})

		// Public invitation token routes (no auth required)
		// These use token-based validation, not JWT
		r.Get("/invitations/{token}", invitationsHandler.GetInvitationByToken)
		r.Post("/invitations/{token}/accept", invitationsHandler.AcceptInvitation)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.JWTAuth(queries))

			// =================================================================
			// CLI Routes (any authenticated employee)
			// These are self-service endpoints for employees using the CLI
			// =================================================================
			r.Route("/employees/me", func(r chi.Router) {
				// Claude token management
				r.Put("/claude-token", claudeTokensHandler.SetEmployeeClaudeToken)
				r.Delete("/claude-token", claudeTokensHandler.DeleteEmployeeClaudeToken)
				r.Get("/claude-token/status", claudeTokensHandler.GetClaudeTokenStatus)
				r.Get("/claude-token/effective", claudeTokensHandler.GetEffectiveClaudeToken)
			})

			// =================================================================
			// Admin Routes (admin only)
			// These are high-privilege endpoints for organization admins
			// =================================================================
			r.Group(func(r chi.Router) {
				r.Use(authmiddleware.RequireRole(queries, "admin"))

				// Roles routes - admin only
				r.Route("/roles", func(r chi.Router) {
					r.Get("/", rolesHandler.ListRoles)
					r.Post("/", rolesHandler.CreateRole)
					r.Get("/{role_id}", rolesHandler.GetRole)
					r.Patch("/{role_id}", rolesHandler.UpdateRole)
					r.Delete("/{role_id}", rolesHandler.DeleteRole)
				})
			})

			// =================================================================
			// Manager Routes (admin or manager)
			// These are management endpoints for team leads and admins
			// =================================================================
			r.Group(func(r chi.Router) {
				r.Use(authmiddleware.RequireRole(queries, "admin", "manager"))

				// Employees routes (list, create, get, update, delete)
				r.Route("/employees", func(r chi.Router) {
					r.Get("/", employeesHandler.ListEmployees)
					r.Post("/", employeesHandler.CreateEmployee)
					r.Get("/{employee_id}", employeesHandler.GetEmployee)
					r.Patch("/{employee_id}", employeesHandler.UpdateEmployee)
					r.Delete("/{employee_id}", employeesHandler.DeleteEmployee)
				})

				// Teams routes
				r.Route("/teams", func(r chi.Router) {
					r.Get("/", teamsHandler.ListTeams)
					r.Post("/", teamsHandler.CreateTeam)
					r.Get("/{team_id}", teamsHandler.GetTeam)
					r.Patch("/{team_id}", teamsHandler.UpdateTeam)
					r.Delete("/{team_id}", teamsHandler.DeleteTeam)
				})
			})

			// Protected invitation routes (require authentication)
			r.Route("/invitations", func(r chi.Router) {
				// List all invitations for organization (admin)
				r.Get("/", invitationsHandler.ListInvitations)
				// Create new invitation (admin)
				r.Post("/", invitationsHandler.CreateInvitation)
				// Cancel invitation (admin)
				r.Delete("/{id}", invitationsHandler.CancelInvitation)
			})

			// Organizations routes
			r.Route("/organizations/current", func(r chi.Router) {
				r.Get("/", organizationsHandler.GetCurrentOrganization)
				r.Patch("/", organizationsHandler.UpdateCurrentOrganization)

				// Organization Claude token (hybrid auth)
				r.Put("/claude-token", claudeTokensHandler.SetOrganizationClaudeToken)
				r.Delete("/claude-token", claudeTokensHandler.DeleteOrganizationClaudeToken)
			})

			// Skills catalog routes
			r.Route("/skills", func(r chi.Router) {
				r.Get("/", skillsHandler.ListSkills)
				r.Get("/{skill_id}", skillsHandler.GetSkill)
			})

			// Employee skills routes
			r.Route("/employees/me/skills", func(r chi.Router) {
				r.Get("/", skillsHandler.ListEmployeeSkills)
				r.Get("/{skill_id}", skillsHandler.GetEmployeeSkill)
			})

			// Employee tool policies routes
			r.Route("/employees/me/tool-policies", func(r chi.Router) {
				r.Get("/", toolPoliciesHandler.GetEmployeeToolPolicies)
			})

			// Activity logs routes (for web UI dashboard)
			r.Route("/activity-logs", func(r chi.Router) {
				r.Get("/", activityLogsHandler.ListActivityLogs)
			})

			// Logging API routes (for CLI and programmatic access)
			r.Route("/logs", func(r chi.Router) {
				r.Post("/", logsHandler.CreateLog)
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					// Extract query parameters
					params := extractListLogsParams(r)
					logsHandler.ListLogs(w, r, params)
				})
				r.Get("/export", func(w http.ResponseWriter, r *http.Request) {
					// Extract query parameters
					params := extractExportLogsParams(r)
					logsHandler.ExportLogs(w, r, params)
				})
				r.Get("/sessions", func(w http.ResponseWriter, r *http.Request) {
					// Extract query parameters
					params := extractListSessionsParams(r)
					logsHandler.ListSessions(w, r, params)
				})

				// WebSocket endpoint for real-time log streaming
				// Format: WS /api/v1/logs/stream?session_id=xxx&employee_id=xxx
				// Auth: JWT token required in Authorization header
				r.Get("/stream", wsHandler.ServeHTTP)
			})

			// Webhook destination routes
			r.Route("/webhooks", func(r chi.Router) {
				r.Get("/", webhooksHandler.ListWebhookDestinations)
				r.Post("/", webhooksHandler.CreateWebhookDestination)
				r.Route("/{webhookId}", func(r chi.Router) {
					r.Get("/", webhooksHandler.GetWebhookDestination)
					r.Patch("/", webhooksHandler.UpdateWebhookDestination)
					r.Delete("/", webhooksHandler.DeleteWebhookDestination)
					r.Post("/test", webhooksHandler.TestWebhookDestination)
					r.Get("/deliveries", webhooksHandler.ListWebhookDeliveries)
				})
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

			// Agent requests routes (keeping for future access request workflow)
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

	// Start webhook forwarder worker (processes every 10 seconds)
	webhookForwarderCtx, webhookForwarderCancel := context.WithCancel(context.Background())
	webhookForwarder := service.NewWebhookForwarder(queries)
	go webhookForwarder.StartForwarderWorker(webhookForwarderCtx, 10*time.Second)

	// Start server in goroutine
	go func() {
		log.Printf("🚀 API Server starting on http://localhost:%s", port)
		log.Printf("📝 Health Check: http://localhost:%s/api/v1/health", port)
		log.Printf("📚 API Documentation: http://localhost:%s/api/docs", port)
		log.Printf("🔐 Auth endpoints:")
		log.Printf("   POST http://localhost:%s/api/v1/auth/login", port)
		log.Printf("   POST http://localhost:%s/api/v1/auth/logout", port)
		log.Printf("   GET  http://localhost:%s/api/v1/auth/me", port)
		log.Printf("🌐 Web UI available at http://localhost:3000 (Next.js app)")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Stop webhook forwarder
	webhookForwarderCancel()

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

// extractListLogsParams extracts query parameters for ListLogs
func extractListLogsParams(r *http.Request) api.ListLogsParams {
	params := api.ListLogsParams{}
	query := r.URL.Query()

	// Parse string parameters
	if clientName := query.Get("client_name"); clientName != "" {
		params.ClientName = &clientName
	}

	// Parse UUID parameters
	if sessionID := query.Get("session_id"); sessionID != "" {
		if uid, err := uuid.Parse(sessionID); err == nil {
			apiUUID := openapi_types.UUID(uid)
			params.SessionId = &apiUUID
		}
	}
	if employeeID := query.Get("employee_id"); employeeID != "" {
		if uid, err := uuid.Parse(employeeID); err == nil {
			apiUUID := openapi_types.UUID(uid)
			params.EmployeeId = &apiUUID
		}
	}

	// Parse event type
	if eventType := query.Get("event_type"); eventType != "" {
		et := api.ListLogsParamsEventType(eventType)
		params.EventType = &et
	}

	// Parse event category
	if eventCategory := query.Get("event_category"); eventCategory != "" {
		ec := api.ListLogsParamsEventCategory(eventCategory)
		params.EventCategory = &ec
	}

	// Parse pagination
	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			p := api.Page(page)
			params.Page = &p
		}
	}
	if perPageStr := query.Get("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil {
			pp := api.PerPage(perPage)
			params.PerPage = &pp
		}
	}

	// Parse date filters
	if startDate := query.Get("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			params.StartDate = &t
		}
	}
	if endDate := query.Get("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			params.EndDate = &t
		}
	}

	return params
}

// extractExportLogsParams extracts query parameters for ExportLogs
func extractExportLogsParams(r *http.Request) api.ExportLogsParams {
	params := api.ExportLogsParams{}
	// TODO: Parse query parameters from r.URL.Query()
	// For now, return empty params - will be implemented in follow-up
	return params
}

// extractListSessionsParams extracts query parameters for ListSessions
func extractListSessionsParams(r *http.Request) api.ListSessionsParams {
	params := api.ListSessionsParams{}
	// TODO: Parse query parameters from r.URL.Query()
	// For now, return empty params - will be implemented in follow-up
	return params
}
