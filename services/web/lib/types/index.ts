/**
 * Centralized type exports for the web application.
 *
 * Re-exports types from the generated OpenAPI schema for convenience.
 * Import from '@/lib/types' instead of '@/lib/api/schema' for commonly used types.
 */

import type { components } from '@/lib/api/schema';

// Agent types
export type Agent = components['schemas']['Agent'];
export type OrgAgentConfig = components['schemas']['OrgAgentConfig'];
export type TeamAgentConfig = components['schemas']['TeamAgentConfig'];

// Employee types
export type Employee = components['schemas']['Employee'];

// Team types
export type Team = components['schemas']['Team'];

// Role types
export type Role = components['schemas']['Role'];

// Organization types
export type Organization = components['schemas']['Organization'];
