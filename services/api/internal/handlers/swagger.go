package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// SwaggerHandler returns a handler function that serves Swagger UI
func SwaggerHandler() http.HandlerFunc {
	// Get the Swagger UI handler
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/api/docs/spec.yaml"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)

	// Wrap http.Handler as http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		swaggerHandler.ServeHTTP(w, r)
	}
}

// SpecHandler serves the OpenAPI spec file
func SpecHandler() http.HandlerFunc {
	projectRoot := getProjectRoot()
	specPath := filepath.Join(projectRoot, "shared", "openapi", "spec.yaml")

	return func(w http.ResponseWriter, r *http.Request) {
		// Read the spec file
		content, err := os.ReadFile(specPath)
		if err != nil {
			http.Error(w, "OpenAPI spec not found", http.StatusNotFound)
			return
		}

		// Set content type to YAML
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

// getProjectRoot returns the project root directory
// Assumes we're in services/api/internal/handlers
func getProjectRoot() string {
	// Try to get from environment variable first
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		return root
	}

	// Default: navigate up from binary location
	// This will work when running from the project root or via make commands
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Navigate up to find the project root
	// Look for go.work file which indicates the root
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, fallback to current dir
			return wd
		}
		dir = parent
	}
}
