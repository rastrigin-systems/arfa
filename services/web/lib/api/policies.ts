import { getErrorMessage } from './errors';
import type { components } from './schema';

// Use schema types for API responses
type SchemaToolPolicy = components['schemas']['ToolPolicy'];
type CreateRequest = components['schemas']['CreateToolPolicyRequest'];
type UpdateRequest = components['schemas']['UpdateToolPolicyRequest'];

// Re-export for backwards compatibility
export type ToolPolicy = SchemaToolPolicy;
export type CreateToolPolicyRequest = CreateRequest;
export type UpdateToolPolicyRequest = UpdateRequest;

/**
 * Get list of all tool policies in the organization
 * Uses Next.js API route which handles auth token forwarding
 */
export async function getPolicies(): Promise<ToolPolicy[]> {
  const response = await fetch('/api/policies', {
    credentials: 'include',
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(getErrorMessage(errorData, 'Failed to fetch policies'));
  }

  const data = await response.json();
  return data.policies || [];
}

/**
 * Create a new tool policy
 */
export async function createPolicy(data: CreateToolPolicyRequest): Promise<ToolPolicy> {
  const response = await fetch('/api/policies', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(getErrorMessage(errorData, 'Failed to create policy'));
  }

  return response.json();
}

/**
 * Update an existing tool policy
 */
export async function updatePolicy(id: string, data: UpdateToolPolicyRequest): Promise<ToolPolicy> {
  const response = await fetch(`/api/policies/${id}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(getErrorMessage(errorData, 'Failed to update policy'));
  }

  return response.json();
}

/**
 * Delete a tool policy
 */
export async function deletePolicy(id: string): Promise<void> {
  const response = await fetch(`/api/policies/${id}`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(getErrorMessage(errorData, 'Failed to delete policy'));
  }
}
