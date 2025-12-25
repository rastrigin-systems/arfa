package testutil

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rastrigin-systems/arfa/generated/db"
)

// SetupTestDB creates a PostgreSQL testcontainer and returns a connection with db.Queries
//
// Integration Test Helper: This spins up a REAL PostgreSQL database in Docker
// for each test. It's slower than mocks but tests the real thing!
func SetupTestDB(t *testing.T) (*pgx.Conn, *db.Queries) {
	ctx := context.Background()

	// Get absolute path to schema.sql (in monorepo platform directory)
	schemaPath, err := filepath.Abs("../../../../platform/database/schema.sql")
	require.NoError(t, err)

	// Start PostgreSQL container using generic testcontainers API
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "arfa",
			"POSTGRES_PASSWORD": "test_password",
			"POSTGRES_DB":       "arfa_test",
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      schemaPath,
				ContainerFilePath: "/docker-entrypoint-initdb.d/schema.sql",
				FileMode:          0755,
			},
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start PostgreSQL container")

	// Cleanup container when test finishes
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	// Get container host and port
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Build connection string
	connStr := fmt.Sprintf("postgres://arfa:test_password@%s:%s/arfa_test?sslmode=disable",
		host, port.Port())

	// Connect to database
	conn, err := pgx.Connect(ctx, connStr)
	require.NoError(t, err, "Failed to connect to test database")

	t.Cleanup(func() {
		_ = conn.Close(ctx)
	})

	// Create queries instance
	queries := db.New(conn)

	return conn, queries
}

// GetContext returns a test context with timeout
func GetContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}
