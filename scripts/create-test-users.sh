#!/bin/bash
set -e

# Create test organizations and employees
# Password for all accounts: password123

echo "🔧 Creating test organizations and employees..."

# Generate bcrypt hash for 'password123' (cost 10)
# Using Go to generate a fresh hash
cat > /tmp/genhash.go <<'GO'
package main
import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)
func main() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	fmt.Print(string(hash))
}
GO

HASH=$(go run /tmp/genhash.go)
rm /tmp/genhash.go

echo "Generated password hash: $HASH"

# Apply SQL
docker exec -i ubik-postgres psql -U ubik -d ubik <<SQL
-- ============================================================================
-- TEST DATA: Organizations and Employees
-- ============================================================================
-- All passwords: password123

-- 1. Acme Corp (Mature Enterprise)
INSERT INTO organizations (name, slug, plan, settings, max_employees, max_agents_per_employee)
VALUES ('Acme Corp', 'acme-corp', 'enterprise', '{}', 100, 10)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

DO \$\$
DECLARE
    acme_id UUID;
    admin_role_id UUID;
    manager_role_id UUID;
    dev_role_id UUID;
BEGIN
    SELECT id INTO acme_id FROM organizations WHERE slug = 'acme-corp';
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';
    SELECT id INTO manager_role_id FROM roles WHERE name = 'manager';
    SELECT id INTO dev_role_id FROM roles WHERE name = 'developer';

    -- Sarah Chen (Owner/CTO - Admin role)
    INSERT INTO employees (org_id, email, full_name, password_hash, status, role_id)
    VALUES (acme_id, 'sarah.cto@acme.com', 'Sarah Chen', '$HASH', 'active', admin_role_id)
    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name;

    -- Alex Rodriguez (Manager)
    INSERT INTO employees (org_id, email, full_name, password_hash, status, role_id)
    VALUES (acme_id, 'alex.manager@acme.com', 'Alex Rodriguez', '$HASH', 'active', manager_role_id)
    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name;

    -- Maria Garcia (Senior Developer)
    INSERT INTO employees (org_id, email, full_name, password_hash, status, role_id)
    VALUES (acme_id, 'maria.senior@acme.com', 'Maria Garcia', '$HASH', 'active', dev_role_id)
    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name;

    -- Emma Wilson (Frontend Developer)
    INSERT INTO employees (org_id, email, full_name, password_hash, status, role_id)
    VALUES (acme_id, 'emma.frontend@acme.com', 'Emma Wilson', '$HASH', 'active', dev_role_id)
    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name;

    RAISE NOTICE 'Created Acme Corp employees';
END \$\$;

-- 2. TechCo (Startup)
INSERT INTO organizations (name, slug, plan, settings, max_employees, max_agents_per_employee)
VALUES ('TechCo', 'techco', 'pro', '{}', 50, 5)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name;

DO \$\$
DECLARE
    techco_id UUID;
    admin_role_id UUID;
    dev_role_id UUID;
BEGIN
    SELECT id INTO techco_id FROM organizations WHERE slug = 'techco';
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';
    SELECT id INTO dev_role_id FROM roles WHERE name = 'developer';

    -- Jane Founder (Owner)
    INSERT INTO employees (org_id, email, full_name, password_hash, status, role_id)
    VALUES (techco_id, 'jane.founder@techco.com', 'Jane Founder', '$HASH', 'active', admin_role_id)
    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name;

    -- Tom Developer
    INSERT INTO employees (org_id, email, full_name, password_hash, status, role_id)
    VALUES (techco_id, 'tom.dev@techco.com', 'Tom Developer', '$HASH', 'active', dev_role_id)
    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name;

    RAISE NOTICE 'Created TechCo employees';
END \$\$;

-- 3. NewCorp (Small business)
INSERT INTO organizations (name, slug, plan, settings, max_employees, max_agents_per_employee)
VALUES ('NewCorp', 'newcorp', 'starter', '{}', 10, 3)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name;

DO \$\$
DECLARE
    newcorp_id UUID;
    admin_role_id UUID;
BEGIN
    SELECT id INTO newcorp_id FROM organizations WHERE slug = 'newcorp';
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';

    -- Owner
    INSERT INTO employees (org_id, email, full_name, password_hash, status, role_id)
    VALUES (newcorp_id, 'owner@newcorp.com', 'Owner NewCorp', '$HASH', 'active', admin_role_id)
    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name;

    RAISE NOTICE 'Created NewCorp employees';
END \$\$;

-- 4. Solo Startup
INSERT INTO organizations (name, slug, plan, settings, max_employees, max_agents_per_employee)
VALUES ('Solo Startup', 'solo-startup', 'starter', '{}', 10, 3)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name;

DO \$\$
DECLARE
    solo_id UUID;
    admin_role_id UUID;
BEGIN
    SELECT id INTO solo_id FROM organizations WHERE slug = 'solo-startup';
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';

    -- John Solo
    INSERT INTO employees (org_id, email, full_name, password_hash, status, role_id)
    VALUES (solo_id, 'john@solostartup.com', 'John Solo', '$HASH', 'active', admin_role_id)
    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name;

    RAISE NOTICE 'Created Solo Startup employees';
END \$\$;

SQL

echo "✅ Test users created successfully!"
echo ""
echo "Test credentials (all passwords: 'password123'):"
echo ""
echo "Acme Corp (Mature Enterprise):"
echo "  sarah.cto@acme.com      (Owner/CTO)"
echo "  alex.manager@acme.com   (Manager)"
echo "  maria.senior@acme.com   (Developer)"
echo "  emma.frontend@acme.com  (Developer)"
echo ""
echo "Other Companies:"
echo "  jane.founder@techco.com (TechCo Owner)"
echo "  tom.dev@techco.com      (TechCo Developer)"
echo "  owner@newcorp.com       (NewCorp Owner)"
echo "  john@solostartup.com    (Solo Startup Owner)"
