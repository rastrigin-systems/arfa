/**
 * MSW Handlers for E2E Tests
 *
 * These handlers intercept API calls on both client and server side,
 * replacing the need for a real backend during E2E tests.
 *
 * MSW works by intercepting fetch/XHR requests at the network level,
 * making it work for both browser and Node.js (server components).
 */

import { http, HttpResponse } from 'msw';
import {
  mockOrganization,
  mockEmployees,
  mockTeams,
  mockRoles,
  mockAgents,
  mockOrgAgentConfigs,
  mockTeamAgentConfigs,
  mockEmployeeAgentConfigs,
  mockResolvedAgentConfigs,
  getMockEmployeeByEmail,
  filterEmployeesByStatus,
  searchEmployees,
} from '../fixtures/mockData';

const API_BASE = 'http://localhost:8080/api/v1';

/**
 * All MSW handlers for the API
 */
export const handlers = [
  // ==========================================
  // Authentication Endpoints
  // ==========================================

  // POST /auth/login
  http.post(`${API_BASE}/auth/login`, async ({ request }) => {
    const body = (await request.json()) as { email: string; password: string };

    // Validate credentials
    const employee = getMockEmployeeByEmail(body.email);
    if (!employee || body.password !== 'password123') {
      return HttpResponse.json(
        {
          error: 'unauthorized',
          message: 'Invalid credentials',
        },
        { status: 401 }
      );
    }

    // Return successful login response
    return HttpResponse.json({
      token: 'mock-jwt-token-12345',
      expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
      employee: employee,
    });
  }),

  // POST /auth/logout
  http.post(`${API_BASE}/auth/logout`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  // GET /auth/me
  http.get(`${API_BASE}/auth/me`, () => {
    // Return first employee (Alice) as the logged-in user
    return HttpResponse.json(mockEmployees[0]);
  }),

  // ==========================================
  // Employee Endpoints
  // ==========================================

  // GET /employees - List employees with filtering
  http.get(`${API_BASE}/employees`, ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get('status') as
      | 'active'
      | 'suspended'
      | 'inactive'
      | null;
    const search = url.searchParams.get('search');
    const page = parseInt(url.searchParams.get('page') || '1');
    const per_page = parseInt(url.searchParams.get('per_page') || '20');

    let filteredEmployees = mockEmployees;

    // Filter by status
    if (status) {
      filteredEmployees = filterEmployeesByStatus(status);
    }

    // Filter by search query
    if (search) {
      filteredEmployees = searchEmployees(search);
    }

    // Pagination
    const offset = (page - 1) * per_page;
    const paginatedEmployees = filteredEmployees.slice(offset, offset + per_page);

    return HttpResponse.json({
      employees: paginatedEmployees,
      total: filteredEmployees.length,
      limit: per_page,
      offset: offset,
    });
  }),

  // GET /employees/:id - Get single employee
  http.get(`${API_BASE}/employees/:id`, ({ params }) => {
    const { id } = params;
    const employee = mockEmployees.find((emp) => emp.id === id);

    if (!employee) {
      return HttpResponse.json(
        {
          error: 'not_found',
          message: 'Employee not found',
        },
        { status: 404 }
      );
    }

    return HttpResponse.json(employee);
  }),

  // POST /employees - Create employee
  http.post(`${API_BASE}/employees`, async ({ request }) => {
    const body = (await request.json()) as {
      team_id?: string;
      role_id: string;
      email: string;
      full_name: string;
      preferences?: Record<string, unknown>;
    };

    // Create new employee with mock data
    const newEmployee = {
      id: `employee-new-${Date.now()}`,
      org_id: mockOrganization.id,
      team_id: body.team_id || null,
      team_name: mockTeams.find((t) => t.id === body.team_id)?.name || null,
      role_id: body.role_id,
      email: body.email,
      full_name: body.full_name,
      status: 'active' as const,
      has_personal_claude_token: false,
      preferences: body.preferences || {},
      last_login_at: null,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    return HttpResponse.json(
      {
        employee: newEmployee,
        temporary_password: 'TempPass123!',
      },
      { status: 201 }
    );
  }),

  // PATCH /employees/:id - Update employee
  http.patch(`${API_BASE}/employees/:id`, async ({ params, request }) => {
    const { id } = params;
    const employee = mockEmployees.find((emp) => emp.id === id);

    if (!employee) {
      return HttpResponse.json(
        {
          error: 'not_found',
          message: 'Employee not found',
        },
        { status: 404 }
      );
    }

    const body = (await request.json()) as Partial<typeof employee>;
    const updatedEmployee = {
      ...employee,
      ...body,
      updated_at: new Date().toISOString(),
    };

    return HttpResponse.json(updatedEmployee);
  }),

  // DELETE /employees/:id - Delete employee
  http.delete(`${API_BASE}/employees/:id`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  // ==========================================
  // Organization Endpoints
  // ==========================================

  // GET /organizations/current
  http.get(`${API_BASE}/organizations/current`, () => {
    return HttpResponse.json(mockOrganization);
  }),

  // PUT /organizations/current/claude-token
  http.put(`${API_BASE}/organizations/current/claude-token`, () => {
    return HttpResponse.json({
      success: true,
      message: 'Claude API token updated successfully',
    });
  }),

  // ==========================================
  // Team Endpoints
  // ==========================================

  // GET /teams - List all teams
  http.get(`${API_BASE}/teams`, () => {
    return HttpResponse.json({
      teams: mockTeams,
      total: mockTeams.length,
    });
  }),

  // GET /teams/:id - Get single team
  http.get(`${API_BASE}/teams/:id`, ({ params }) => {
    const { id } = params;
    const team = mockTeams.find((t) => t.id === id);

    if (!team) {
      return HttpResponse.json(
        {
          error: 'not_found',
          message: 'Team not found',
        },
        { status: 404 }
      );
    }

    return HttpResponse.json(team);
  }),

  // POST /teams - Create team
  http.post(`${API_BASE}/teams`, async ({ request }) => {
    const body = (await request.json()) as {
      name: string;
      description?: string;
    };

    const newTeam = {
      id: `team-new-${Date.now()}`,
      org_id: mockOrganization.id,
      name: body.name,
      description: body.description || null,
      member_count: 0,
      agent_config_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    return HttpResponse.json(newTeam, { status: 201 });
  }),

  // ==========================================
  // Role Endpoints
  // ==========================================

  // GET /roles - List all roles
  http.get(`${API_BASE}/roles`, () => {
    return HttpResponse.json({
      data: mockRoles,
    });
  }),

  // ==========================================
  // Agent Catalog Endpoints
  // ==========================================

  // GET /agents - List available agents
  http.get(`${API_BASE}/agents`, () => {
    return HttpResponse.json({
      agents: mockAgents,
      total: mockAgents.length,
    });
  }),

  // ==========================================
  // Agent Configuration Endpoints
  // ==========================================

  // GET /organizations/current/agent-configs
  http.get(`${API_BASE}/organizations/current/agent-configs`, () => {
    return HttpResponse.json({
      configs: mockOrgAgentConfigs,
      total: mockOrgAgentConfigs.length,
    });
  }),

  // POST /organizations/current/agent-configs
  http.post(`${API_BASE}/organizations/current/agent-configs`, async ({ request }) => {
    const body = (await request.json()) as {
      agent_id: string;
      is_enabled?: boolean;
      config: Record<string, unknown>;
    };

    const agent = mockAgents.find((a) => a.id === body.agent_id);

    const newConfig = {
      id: `org-config-new-${Date.now()}`,
      org_id: mockOrganization.id,
      agent_id: body.agent_id,
      agent_name: agent?.name || 'Unknown',
      agent_type: agent?.type || 'unknown',
      agent_provider: agent?.provider || 'unknown',
      is_enabled: body.is_enabled ?? true,
      config: body.config,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    return HttpResponse.json(newConfig, { status: 201 });
  }),

  // GET /teams/:team_id/agent-configs
  http.get(`${API_BASE}/teams/:teamId/agent-configs`, () => {
    return HttpResponse.json({
      configs: mockTeamAgentConfigs,
      total: mockTeamAgentConfigs.length,
    });
  }),

  // GET /employees/:employee_id/agent-configs
  http.get(`${API_BASE}/employees/:employeeId/agent-configs`, () => {
    return HttpResponse.json({
      configs: mockEmployeeAgentConfigs,
      total: mockEmployeeAgentConfigs.length,
    });
  }),

  // GET /employees/:employee_id/agent-configs/resolved
  http.get(`${API_BASE}/employees/:employeeId/agent-configs/resolved`, () => {
    return HttpResponse.json({
      configs: mockResolvedAgentConfigs,
      total: mockResolvedAgentConfigs.length,
    });
  }),
];
