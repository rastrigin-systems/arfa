-- Migration: Add Claude API Token Support (Hybrid Auth Model)
-- Version: 004
-- Description: Add company-wide and per-employee Claude API tokens

-- Add company-wide Claude token to organizations
ALTER TABLE organizations
ADD COLUMN claude_api_token TEXT; -- Encrypted token, nullable (optional)

COMMENT ON COLUMN organizations.claude_api_token IS
'Company-wide Claude API token (from claude setup-token). Used as default for all employees who dont have personal tokens.';

-- Add personal Claude token to employees
ALTER TABLE employees
ADD COLUMN personal_claude_token TEXT; -- Encrypted token, nullable (optional)

COMMENT ON COLUMN employees.personal_claude_token IS
'Employee personal Claude API token. If set, takes precedence over organization token.';

-- Add token source tracking to usage records
ALTER TABLE usage_records
ADD COLUMN token_source VARCHAR(20) DEFAULT 'company' CHECK (token_source IN ('company', 'personal'));

COMMENT ON COLUMN usage_records.token_source IS
'Indicates which token was used: company (org token) or personal (employee token)';

-- Create index for token status queries
CREATE INDEX idx_employees_personal_token ON employees(org_id, personal_claude_token)
WHERE personal_claude_token IS NOT NULL;

-- Add function to get effective token for employee
CREATE OR REPLACE FUNCTION get_effective_claude_token(emp_id UUID)
RETURNS TABLE (
    token TEXT,
    source VARCHAR(20),
    org_id UUID
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(e.personal_claude_token, o.claude_api_token) as token,
        CASE
            WHEN e.personal_claude_token IS NOT NULL THEN 'personal'::VARCHAR(20)
            ELSE 'company'::VARCHAR(20)
        END as source,
        e.org_id
    FROM employees e
    JOIN organizations o ON e.org_id = o.id
    WHERE e.id = emp_id
    AND e.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_effective_claude_token(UUID) IS
'Returns the effective Claude token for an employee (personal if available, otherwise company token)';

-- Rollback script (for reference)
/*
DROP FUNCTION IF EXISTS get_effective_claude_token(UUID);
DROP INDEX IF EXISTS idx_employees_personal_token;
ALTER TABLE usage_records DROP COLUMN IF EXISTS token_source;
ALTER TABLE employees DROP COLUMN IF EXISTS personal_claude_token;
ALTER TABLE organizations DROP COLUMN IF EXISTS claude_api_token;
*/
