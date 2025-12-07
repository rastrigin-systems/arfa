-- Seed Data: Gemini CLI Configuration
-- Description: Adds Gemini CLI agent to the catalog

-- Agent: Gemini CLI
INSERT INTO agents (name, type, description, provider, llm_provider, llm_model, default_config, capabilities)
VALUES (
    'Gemini CLI',
    'gemini-cli',
    'Google Gemini CLI Agent',
    'google',
    'google',
    'gemini-2.0-flash',
    '{"temperature": 0.7}'::JSONB,
    '["code_generation", "debugging", "refactoring", "research"]'::JSONB
)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    llm_model = EXCLUDED.llm_model,
    updated_at = NOW();
