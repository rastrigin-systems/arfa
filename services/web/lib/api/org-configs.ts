/**
 * Organization Agent Configs API
 *
 * Type-safe API calls for organization-level agent configurations.
 *
 * Server-side functions (prefixed with nothing) use apiClient directly with auth headers.
 * Client-side functions (prefixed with "client") use fetch to Next.js API routes.
 */

import { apiClient } from './client';
import { type ApiError, getErrorMessage } from './errors';
import type { components } from './schema';

// Re-export types from schema
export type OrgAgentConfig = components['schemas']['OrgAgentConfig'];
export type CreateOrgAgentConfigRequest = components['schemas']['CreateOrgAgentConfigRequest'];
export type UpdateOrgAgentConfigRequest = components['schemas']['UpdateOrgAgentConfigRequest'];
export type ListOrgAgentConfigsResponse = components['schemas']['ListOrgAgentConfigsResponse'];

// Re-export ApiError for backwards compatibility
export type { ApiError };

// =============================================================================
// Client-side functions (for use in client components via Next.js API routes)
// =============================================================================

/**
 * Create a new org agent config (client-side)
 * Uses Next.js API route to handle auth
 */
export async function clientCreateOrgAgentConfig(
  request: CreateOrgAgentConfigRequest
): Promise<OrgAgentConfig> {
  const response = await fetch('/api/organizations/current/agent-configs', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to create org agent config');
  }

  return response.json();
}

/**
 * Update an org agent config (client-side)
 * Uses Next.js API route to handle auth
 */
export async function clientUpdateOrgAgentConfig(
  configId: string,
  request: UpdateOrgAgentConfigRequest
): Promise<OrgAgentConfig> {
  const response = await fetch(`/api/organizations/current/agent-configs/${configId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to update org agent config');
  }

  return response.json();
}

/**
 * Delete an org agent config (client-side)
 * Uses Next.js API route to handle auth
 */
export async function clientDeleteOrgAgentConfig(configId: string): Promise<void> {
  const response = await fetch(`/api/organizations/current/agent-configs/${configId}`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to delete org agent config');
  }
}

// =============================================================================
// Server-side functions (for use in server components and API routes)
// =============================================================================

/**
 * List all organization agent configurations
 * GET /organizations/current/agent-configs
 */
export async function listOrgAgentConfigs(): Promise<OrgAgentConfig[]> {
  const { data, error, response } = await apiClient.GET('/organizations/current/agent-configs');

  if (error) {
    throw new Error(getErrorMessage(error, `Failed to list org agent configs (${response?.status})`));
  }

  return data?.configs || [];
}

/**
 * Get a single organization agent configuration
 * GET /organizations/current/agent-configs/{config_id}
 */
export async function getOrgAgentConfig(configId: string): Promise<OrgAgentConfig> {
  const { data, error, response } = await apiClient.GET('/organizations/current/agent-configs/{config_id}', {
    params: { path: { config_id: configId } },
  });

  if (error || !data) {
    throw new Error(getErrorMessage(error, `Failed to get org agent config (${response?.status})`));
  }

  return data;
}

/**
 * Create a new organization agent configuration
 * POST /organizations/current/agent-configs
 */
export async function createOrgAgentConfig(
  request: CreateOrgAgentConfigRequest
): Promise<OrgAgentConfig> {
  const { data, error, response } = await apiClient.POST('/organizations/current/agent-configs', {
    body: request,
  });

  if (error || !data) {
    throw new Error(getErrorMessage(error, `Failed to create org agent config (${response?.status})`));
  }

  return data;
}

/**
 * Update an existing organization agent configuration
 * PATCH /organizations/current/agent-configs/{config_id}
 */
export async function updateOrgAgentConfig(
  configId: string,
  request: UpdateOrgAgentConfigRequest
): Promise<OrgAgentConfig> {
  const { data, error, response } = await apiClient.PATCH('/organizations/current/agent-configs/{config_id}', {
    params: { path: { config_id: configId } },
    body: request,
  });

  if (error || !data) {
    throw new Error(getErrorMessage(error, `Failed to update org agent config (${response?.status})`));
  }

  return data;
}

/**
 * Delete an organization agent configuration
 * DELETE /organizations/current/agent-configs/{config_id}
 */
export async function deleteOrgAgentConfig(configId: string): Promise<void> {
  const { error, response } = await apiClient.DELETE('/organizations/current/agent-configs/{config_id}', {
    params: { path: { config_id: configId } },
  });

  if (error) {
    throw new Error(getErrorMessage(error, `Failed to delete org agent config (${response?.status})`));
  }
}
