#!/bin/bash
# Query Ubik Enterprise production database via Cloud SQL Proxy
#
# Usage:
#   ./query_prod.sh "SELECT * FROM employees LIMIT 5;"
#   ./query_prod.sh < query.sql
#   ./query_prod.sh << 'EOF'
#   SELECT * FROM employees;
#   EOF

set -e

# Configuration
PROJECT="ubik-enterprise-prod"
INSTANCE="ubik-postgres"
DATABASE="ubik"
USER="ubik"
PORT=9473

# Paths
PSQL="/opt/homebrew/opt/libpq/bin/psql"
GCLOUD_SDK="/opt/homebrew/share/google-cloud-sdk/bin"
PROXY="$GCLOUD_SDK/cloud-sql-proxy"

# Check dependencies
if [ ! -f "$PSQL" ]; then
    echo "Error: psql not found at $PSQL" >&2
    echo "Install with: brew install libpq" >&2
    exit 1
fi

if [ ! -f "$PROXY" ]; then
    echo "Error: cloud-sql-proxy not found at $PROXY" >&2
    echo "Install with: gcloud components install cloud-sql-proxy" >&2
    exit 1
fi

# Kill any existing proxy on our port
pkill -f "cloud-sql-proxy.*:$PORT" 2>/dev/null || true
sleep 1

# Get password from Secret Manager
PASSWORD=$(PATH="$GCLOUD_SDK:$PATH" gcloud secrets versions access latest \
    --secret=database-url \
    --project=$PROJECT 2>/dev/null | \
    sed -n 's/.*:\/\/ubik:\([^@]*\)@.*/\1/p')

if [ -z "$PASSWORD" ]; then
    echo "Error: Could not retrieve database password from Secret Manager" >&2
    exit 1
fi

# Start proxy in background
$PROXY "$PROJECT:us-central1:$INSTANCE" --port=$PORT &
PROXY_PID=$!

# Wait for proxy to be ready
sleep 3

# Cleanup function
cleanup() {
    kill $PROXY_PID 2>/dev/null || true
}
trap cleanup EXIT

# Run query
if [ -n "$1" ]; then
    # Query passed as argument
    PGPASSWORD="$PASSWORD" $PSQL -h 127.0.0.1 -p $PORT -U $USER -d $DATABASE -c "$1"
elif [ ! -t 0 ]; then
    # Query from stdin
    PGPASSWORD="$PASSWORD" $PSQL -h 127.0.0.1 -p $PORT -U $USER -d $DATABASE
else
    # Interactive mode
    echo "Connected to $DATABASE@$INSTANCE"
    PGPASSWORD="$PASSWORD" $PSQL -h 127.0.0.1 -p $PORT -U $USER -d $DATABASE
fi
