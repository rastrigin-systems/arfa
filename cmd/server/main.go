package main

import (
	"context"
	"fmt"
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
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
	authmiddleware "github.com/sergeirastrigin/ubik-enterprise/internal/middleware"
)

func main() {
	// Get configuration from environment
	dbURL := getEnv("DATABASE_URL", "postgres://pivot:pivot_dev_password@localhost:5432/pivot?sslmode=disable")
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
	authHandler := handlers.NewAuthHandler(queries)
	employeesHandler := handlers.NewEmployeesHandler(queries)
	teamsHandler := handlers.NewTeamsHandler(queries)
	orgAgentConfigsHandler := handlers.NewOrgAgentConfigsHandler(queries)
	teamAgentConfigsHandler := handlers.NewTeamAgentConfigsHandler(queries)
	employeeAgentConfigsHandler := handlers.NewEmployeeAgentConfigsHandler(queries)

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

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
		})

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
			r.Route("/organizations/current/agent-configs", func(r chi.Router) {
				r.Get("/", orgAgentConfigsHandler.ListOrgAgentConfigs)
				r.Post("/", orgAgentConfigsHandler.CreateOrgAgentConfig)
				r.Get("/{config_id}", orgAgentConfigsHandler.GetOrgAgentConfig)
				r.Patch("/{config_id}", orgAgentConfigsHandler.UpdateOrgAgentConfig)
				r.Delete("/{config_id}", orgAgentConfigsHandler.DeleteOrgAgentConfig)
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
		})

		// TODO: Add more routes as they are implemented
		// - /roles (CRUD)
		// - /agents (catalog - read-only)
		// - /mcps (catalog, configs)
		// - /approvals (workflow)
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
