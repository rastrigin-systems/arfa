# Quick Start Guide 🚀

## Prerequisites

Ensure you have installed:
- ✅ Go 1.24+ (`go version`)
- ✅ Docker Desktop (start it!)
- ✅ Make (`make --version`)

## Setup Steps (5 minutes)

### 1. Start Docker Desktop
```bash
# On macOS
open -a Docker

# Wait for Docker to start (you'll see the whale icon in menu bar)
```

### 2. Initialize the Project
```bash
cd /Users/sergeirastrigin/Projects/ubik/pivot

# Install code generation tools (one-time setup)
make install-tools

# Expected output:
# ✅ oapi-codegen installed
# ✅ sqlc installed  
# ✅ tbls installed
```

### 3. Start PostgreSQL
```bash
make db-up

# Expected output:
# ✅ PostgreSQL is ready
# Database connection: postgres://pivot:pivot_dev_password@localhost:5432/pivot
# Web UI: http://localhost:8080 (Adminer)
```

### 4. Verify Database
```bash
# Check if schema was applied
docker exec -it pivot-postgres psql -U pivot -d pivot -c "\dt"

# Should show 17 tables:
# - organizations
# - employees
# - teams
# - roles
# ... etc
```

### 5. Generate Code
```bash
# Generate ERD from database
make generate-erd

# Generate API code from OpenAPI spec
make generate-api

# Generate database code from SQL queries
make generate-db

# Or generate everything at once:
make generate
```

### 6. Verify Generated Code
```bash
# Check generated files
ls -la generated/api/
ls -la generated/db/
ls -la docs/
```

## What You Have Now

✅ **PostgreSQL running** with 17 tables and seed data
✅ **OpenAPI spec** (`openapi/spec.yaml`) with auth + employee endpoints
✅ **SQL queries** (`sqlc/queries/`) for employees, auth, organizations
✅ **Configuration files** for oapi-codegen and sqlc
✅ **Go module** initialized (`go.mod`)

## Next Steps

### Option A: Build Your First Handler

Create the authentication handler:

```bash
# Create handler file
cat > internal/handlers/auth.go << 'EOF'
package handlers

import (
    "net/http"
    "github.com/sergeirastrigin/ubik-enterprise/generated/api"
)

type AuthHandler struct {
    // TODO: Add dependencies (DB, JWT service, etc.)
}

func NewAuthHandler() *AuthHandler {
    return &AuthHandler{}
}

// Login implements the POST /auth/login endpoint
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req api.LoginRequest
    // TODO: Decode request
    // TODO: Validate credentials
    // TODO: Generate JWT token
    // TODO: Return LoginResponse
    
    w.WriteHeader(http.StatusOK)
}
EOF

# Now implement the TODOs!
```

### Option B: Run Example Tests

Create a test to verify generated code works:

```bash
cat > internal/handlers/auth_test.go << 'EOF'
package handlers

import (
    "testing"
)

func TestAuthHandler(t *testing.T) {
    handler := NewAuthHandler()
    if handler == nil {
        t.Fatal("Expected handler to be created")
    }
}
EOF

# Run test
go test ./internal/handlers/
```

### Option C: Start the API Server

Create the main server:

```bash
cat > cmd/server/main.go << 'EOF'
package main

import (
    "log"
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    log.Println("Starting server on :3001")
    log.Fatal(http.ListenAndServe(":3001", r))
}
EOF

# Install dependencies
go mod tidy

# Run server
go run cmd/server/main.go

# Test it
curl http://localhost:3001/health
```

## Troubleshooting

### Docker not starting
```bash
# Check if Docker daemon is running
docker ps

# If not, start Docker Desktop app
# On macOS: open -a Docker
# Wait 30 seconds for it to fully start
```

### Port 5432 already in use
```bash
# Find what's using port 5432
lsof -i :5432

# Option 1: Stop existing PostgreSQL
# Option 2: Change port in docker-compose.yml to "5433:5432"
```

### Code generation fails
```bash
# Ensure tools are installed
which oapi-codegen
which sqlc
which tbls

# If missing, run:
make install-tools

# Check Go PATH
echo $GOPATH
# Should be set, typically ~/go
```

### Database schema not applied
```bash
# Manually apply schema
docker exec -i pivot-postgres psql -U pivot -d pivot < schema.sql

# Verify tables exist
docker exec -it pivot-postgres psql -U pivot -d pivot -c "\dt"
```

## Useful Commands

```bash
# Database
make db-up              # Start PostgreSQL
make db-down            # Stop PostgreSQL  
make db-reset           # Reset database (⚠️ deletes all data)

# Code Generation
make generate           # Generate everything
make generate-erd       # Generate ERD only
make generate-api       # Generate API code only
make generate-db        # Generate DB code only

# Development
make dev                # Start dev server with live reload (once implemented)
make test               # Run tests
make build              # Build binaries
make clean              # Clean generated files

# Help
make help               # Show all commands
```

## Access URLs

- **PostgreSQL**: `localhost:5432`
- **Adminer (DB UI)**: http://localhost:8080
  - System: PostgreSQL
  - Server: postgres
  - Username: pivot
  - Password: pivot_dev_password
  - Database: pivot

- **API Server** (once running): http://localhost:3001
- **Health Check**: http://localhost:3001/health

## File Structure

```
pivot/
├── openapi/
│   ├── spec.yaml              ✅ OpenAPI 3.1 spec
│   └── oapi-codegen.yaml      ✅ Generator config
│
├── sqlc/
│   ├── sqlc.yaml              ✅ Generator config
│   └── queries/
│       ├── employees.sql      ✅ Employee CRUD
│       ├── auth.sql           ✅ Session management
│       └── organizations.sql  ✅ Org/team queries
│
├── generated/                 🤖 Auto-generated (don't edit!)
│   ├── api/                   ← From OpenAPI spec
│   └── db/                    ← From SQL queries
│
├── internal/
│   ├── handlers/              📝 Your code goes here
│   ├── service/               📝 Business logic
│   └── middleware/            📝 Auth, logging, etc.
│
├── cmd/
│   ├── server/                📝 API server entrypoint
│   └── cli/                   📝 Employee CLI (future)
│
├── docs/                      📊 Auto-generated ERD
├── schema.sql                 ✅ PostgreSQL schema
├── Makefile                   ✅ Automation
└── docker-compose.yml         ✅ Local environment
```

## What's Already Done

✅ Database schema (17 tables)
✅ OpenAPI spec (auth + employees)
✅ SQL queries (employees, auth, orgs)
✅ Code generation configs
✅ Makefile automation
✅ Docker Compose setup
✅ Go module initialized

## What You Need to Build

📝 Handler implementations
📝 Business logic services
📝 JWT authentication
📝 Middleware (auth, RLS, CORS)
📝 Tests
📝 Main server setup

## Ready to Code!

Your development environment is fully configured. Start with:

1. `make db-up` - Start database
2. `make generate` - Generate code
3. Implement handlers in `internal/handlers/`
4. Create `cmd/server/main.go`
5. `go run cmd/server/main.go`

Good luck! 🚀

---

**Need Help?**
- See [MIGRATION_PLAN.md](./MIGRATION_PLAN.md) for complete roadmap
- See [DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md) for table details
- Run `make help` for all commands
