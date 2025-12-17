/**
 * Agent configuration types.
 *
 * Agent configs are flexible JSON objects that vary by agent type.
 * These types provide documentation while maintaining compatibility
 * with the OpenAPI-generated schema types.
 *
 * Note: These types use `unknown` to remain compatible with the schema
 * types generated from the OpenAPI spec. The schema is the source of truth.
 */

/**
 * Possible values in an agent configuration.
 * Uses `unknown` for compatibility with OpenAPI-generated schema types.
 *
 * In practice, values are typically:
 * - string (e.g., model identifiers, API endpoints)
 * - number (e.g., temperature, max_tokens, rate limits)
 * - boolean (e.g., feature flags)
 * - null
 * - arrays of the above
 * - nested objects
 */
export type AgentConfigValue = unknown;

/**
 * Agent configuration object.
 * A flexible JSON-compatible object for agent settings.
 *
 * Common properties (not enforced at type level):
 * - model: string - LLM model identifier (e.g., "claude-3-5-sonnet-20241022")
 * - temperature: number - Sampling temperature (0-1)
 * - max_tokens: number - Maximum response tokens
 * - api_endpoint: string - Custom API endpoint
 * - features: object - Feature flags
 * - rate_limit: number - Requests per day
 * - cost_limit: number - Monthly cost limit
 *
 * @example
 * const config: AgentConfig = {
 *   model: "claude-3-5-sonnet-20241022",
 *   temperature: 0.2,
 *   max_tokens: 8192,
 *   features: { autocomplete: true }
 * };
 */
export type AgentConfig = { [key: string]: AgentConfigValue };

/**
 * Employee preferences object.
 * Similar to AgentConfig but for user preferences.
 *
 * Common properties (not enforced at type level):
 * - theme: string - UI theme preference
 * - notifications: object - Notification settings
 * - default_agent: string - Default agent to use
 */
export type EmployeePreferences = { [key: string]: AgentConfigValue };
