/**
 * Team Agent Configs API
 *
 * Type-safe API calls for team-level agent configurations.
 *
 * Server-side functions (prefixed with nothing) use apiClient directly with auth headers.
 * Client-side functions (prefixed with "client") use fetch to Next.js API routes.
 */

import { apiClient } from './client';
import { type ApiError, getErrorMessage } from './errors';
import type { components } from './schema';

// Re-export types from schema
export type TeamAgentConfig = components['schemas']['TeamAgentConfig'];
export type CreateTeamAgentConfigRequest = components['schemas']['CreateTeamAgentConfigRequest'];
export type UpdateTeamAgentConfigRequest = components['schemas']['UpdateTeamAgentConfigRequest'];
export type ListTeamAgentConfigsResponse = components['schemas']['ListTeamAgentConfigsResponse'];

// Re-export ApiError for backwards compatibility
export type { ApiError };

// =============================================================================
// Client-side functions (for use in client components via Next.js API routes)
// =============================================================================

/**
 * List team agent configs (client-side)
 * Uses Next.js API route to handle auth
 */
export async function clientListTeamAgentConfigs(teamId: string): Promise<TeamAgentConfig[]> {
  const response = await fetch(`/api/teams/${teamId}/agent-configs`, {
    credentials: 'include',
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to list team agent configs');
  }

  const data = await response.json();
  return data.configs || [];
}

/**
 * Create a new team agent config (client-side)
 * Uses Next.js API route to handle auth
 */
export async function clientCreateTeamAgentConfig(
  teamId: string,
  request: CreateTeamAgentConfigRequest
): Promise<TeamAgentConfig> {
  const response = await fetch(`/api/teams/${teamId}/agent-configs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to create team agent config');
  }

  return response.json();
}

/**
 * Update a team agent config (client-side)
 * Uses Next.js API route to handle auth
 */
export async function clientUpdateTeamAgentConfig(
  teamId: string,
  configId: string,
  request: UpdateTeamAgentConfigRequest
): Promise<TeamAgentConfig> {
  const response = await fetch(`/api/teams/${teamId}/agent-configs/${configId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to update team agent config');
  }

  return response.json();
}

/**
 * Delete a team agent config (client-side)
 * Uses Next.js API route to handle auth
 */
export async function clientDeleteTeamAgentConfig(teamId: string, configId: string): Promise<void> {
  const response = await fetch(`/api/teams/${teamId}/agent-configs/${configId}`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to delete team agent config');
  }
}

// =============================================================================
// Server-side functions (for use in server components and API routes)
// =============================================================================

/**
 * List all team agent configurations
 * GET /teams/{team_id}/agent-configs
 */
export async function listTeamAgentConfigs(teamId: string): Promise<TeamAgentConfig[]> {
  const { data, error, response } = await apiClient.GET('/teams/{team_id}/agent-configs', {
    params: { path: { team_id: teamId } },
  });

  if (error) {
    throw new Error(getErrorMessage(error, `Failed to list team agent configs (${response?.status})`));
  }

  return data?.configs || [];
}

/**
 * Get a single team agent configuration
 * GET /teams/{team_id}/agent-configs/{config_id}
 */
export async function getTeamAgentConfig(teamId: string, configId: string): Promise<TeamAgentConfig> {
  const { data, error, response } = await apiClient.GET('/teams/{team_id}/agent-configs/{config_id}', {
    params: { path: { team_id: teamId, config_id: configId } },
  });

  if (error || !data) {
    throw new Error(getErrorMessage(error, `Failed to get team agent config (${response?.status})`));
  }

  return data;
}

/**
 * Create a new team agent configuration
 * POST /teams/{team_id}/agent-configs
 */
export async function createTeamAgentConfig(
  teamId: string,
  request: CreateTeamAgentConfigRequest
): Promise<TeamAgentConfig> {
  const { data, error, response } = await apiClient.POST('/teams/{team_id}/agent-configs', {
    params: { path: { team_id: teamId } },
    body: request,
  });

  if (error || !data) {
    throw new Error(getErrorMessage(error, `Failed to create team agent config (${response?.status})`));
  }

  return data;
}

/**
 * Update an existing team agent configuration
 * PATCH /teams/{team_id}/agent-configs/{config_id}
 */
export async function updateTeamAgentConfig(
  teamId: string,
  configId: string,
  request: UpdateTeamAgentConfigRequest
): Promise<TeamAgentConfig> {
  const { data, error, response } = await apiClient.PATCH('/teams/{team_id}/agent-configs/{config_id}', {
    params: { path: { team_id: teamId, config_id: configId } },
    body: request,
  });

  if (error || !data) {
    throw new Error(getErrorMessage(error, `Failed to update team agent config (${response?.status})`));
  }

  return data;
}

/**
 * Delete a team agent configuration
 * DELETE /teams/{team_id}/agent-configs/{config_id}
 */
export async function deleteTeamAgentConfig(teamId: string, configId: string): Promise<void> {
  const { error, response } = await apiClient.DELETE('/teams/{team_id}/agent-configs/{config_id}', {
    params: { path: { team_id: teamId, config_id: configId } },
  });

  if (error) {
    throw new Error(getErrorMessage(error, `Failed to delete team agent config (${response?.status})`));
  }
}
