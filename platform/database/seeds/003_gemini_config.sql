-- Seed Data: Gemini CLI Configuration
-- Description: Adds Gemini CLI agent to the catalog

-- Agent: Gemini CLI
INSERT INTO agents (name, type, description, provider, docker_image, llm_provider, llm_model, default_config, capabilities)
VALUES (
    'Gemini CLI',
    'gemini',
    'Google Gemini CLI Agent',
    'google',
    'ubik/gemini:latest',
    'google',
    'gemini-1.0-pro',
    '{"temperature": 0.7}'::JSONB,
    '["code_generation", "debugging", "refactoring", "research"]'::JSONB
)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    docker_image = EXCLUDED.docker_image,
    llm_model = EXCLUDED.llm_model,
    updated_at = NOW();
