/**
 * API Route Mocking for E2E Tests
 *
 * This file contains Playwright route handlers that intercept API calls
 * and return mock data, eliminating the need for a running backend.
 */

import { Page, Route } from '@playwright/test';
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
} from './mockData';

const API_BASE = '/api/v1';

/**
 * Set up all API route mocks for testing
 */
export async function setupApiMocks(page: Page): Promise<void> {
  // Mock authentication endpoints
  await mockAuthEndpoints(page);

  // Mock employee endpoints
  await mockEmployeeEndpoints(page);

  // Mock organization endpoints
  await mockOrganizationEndpoints(page);

  // Mock team endpoints
  await mockTeamEndpoints(page);

  // Mock role endpoints
  await mockRoleEndpoints(page);

  // Mock agent endpoints
  await mockAgentEndpoints(page);

  // Mock agent configuration endpoints
  await mockAgentConfigEndpoints(page);
}

/**
 * Mock authentication endpoints
 */
async function mockAuthEndpoints(page: Page): Promise<void> {
  // POST /auth/login
  await page.route(`${API_BASE}/auth/login`, async (route: Route) => {
    const request = route.request();
    const postData = request.postDataJSON();

    // Validate credentials
    const employee = getMockEmployeeByEmail(postData.email);
    if (!employee || postData.password !== 'password123') {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'unauthorized',
          message: 'Invalid credentials',
        }),
      });
      return;
    }

    // Return successful login response
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        token: 'mock-jwt-token-12345',
        expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
        employee: employee,
      }),
    });
  });

  // POST /auth/logout
  await page.route(`${API_BASE}/auth/logout`, async (route: Route) => {
    await route.fulfill({
      status: 204,
    });
  });

  // GET /auth/me
  await page.route(`${API_BASE}/auth/me`, async (route: Route) => {
    // Return first employee (Alice) as the logged-in user
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(mockEmployees[0]),
    });
  });
}

/**
 * Mock employee endpoints
 */
async function mockEmployeeEndpoints(page: Page): Promise<void> {
  // GET /employees - List employees with filtering
  await page.route(`${API_BASE}/employees**`, async (route: Route) => {
    const url = new URL(route.request().url());
    const status = url.searchParams.get('status') as
      | 'active'
      | 'suspended'
      | 'inactive'
      | null;
    const search = url.searchParams.get('search');
    const page_num = parseInt(url.searchParams.get('page') || '1');
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
    const offset = (page_num - 1) * per_page;
    const paginatedEmployees = filteredEmployees.slice(offset, offset + per_page);

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        employees: paginatedEmployees,
        total: filteredEmployees.length,
        limit: per_page,
        offset: offset,
      }),
    });
  });

  // GET /employees/:id - Get single employee
  await page.route(
    new RegExp(`${API_BASE}/employees/[a-f0-9-]+$`),
    async (route: Route) => {
      const employeeId = route.request().url().split('/').pop();
      const employee = mockEmployees.find((emp) => emp.id === employeeId);

      if (!employee) {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'not_found',
            message: 'Employee not found',
          }),
        });
        return;
      }

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(employee),
      });
    }
  );

  // POST /employees - Create employee
  await page.route(`${API_BASE}/employees`, async (route: Route) => {
    if (route.request().method() !== 'POST') return;

    const postData = route.request().postDataJSON();

    // Create new employee with mock data
    const newEmployee = {
      id: `employee-new-${Date.now()}`,
      org_id: mockOrganization.id,
      team_id: postData.team_id || null,
      team_name: mockTeams.find((t) => t.id === postData.team_id)?.name || null,
      role_id: postData.role_id,
      email: postData.email,
      full_name: postData.full_name,
      status: 'active' as const,
      has_personal_claude_token: false,
      preferences: postData.preferences || {},
      last_login_at: null,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    await route.fulfill({
      status: 201,
      contentType: 'application/json',
      body: JSON.stringify({
        employee: newEmployee,
        temporary_password: 'TempPass123!',
      }),
    });
  });

  // PATCH /employees/:id - Update employee
  await page.route(
    new RegExp(`${API_BASE}/employees/[a-f0-9-]+$`),
    async (route: Route) => {
      if (route.request().method() !== 'PATCH') return;

      const employeeId = route.request().url().split('/').pop();
      const employee = mockEmployees.find((emp) => emp.id === employeeId);

      if (!employee) {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'not_found',
            message: 'Employee not found',
          }),
        });
        return;
      }

      const patchData = route.request().postDataJSON();
      const updatedEmployee = {
        ...employee,
        ...patchData,
        updated_at: new Date().toISOString(),
      };

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(updatedEmployee),
      });
    }
  );

  // DELETE /employees/:id - Delete employee
  await page.route(
    new RegExp(`${API_BASE}/employees/[a-f0-9-]+$`),
    async (route: Route) => {
      if (route.request().method() !== 'DELETE') return;

      await route.fulfill({
        status: 204,
      });
    }
  );
}

/**
 * Mock organization endpoints
 */
async function mockOrganizationEndpoints(page: Page): Promise<void> {
  // GET /organizations/current
  await page.route(
    `${API_BASE}/organizations/current`,
    async (route: Route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockOrganization),
      });
    }
  );

  // PUT /organizations/current/claude-token
  await page.route(
    `${API_BASE}/organizations/current/claude-token`,
    async (route: Route) => {
      if (route.request().method() !== 'PUT') return;

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Claude API token updated successfully',
        }),
      });
    }
  );
}

/**
 * Mock team endpoints
 */
async function mockTeamEndpoints(page: Page): Promise<void> {
  // GET /teams - List all teams
  await page.route(`${API_BASE}/teams`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        teams: mockTeams,
        total: mockTeams.length,
      }),
    });
  });

  // GET /teams/:id - Get single team
  await page.route(
    new RegExp(`${API_BASE}/teams/[a-f0-9-]+$`),
    async (route: Route) => {
      const teamId = route.request().url().split('/').pop();
      const team = mockTeams.find((t) => t.id === teamId);

      if (!team) {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'not_found',
            message: 'Team not found',
          }),
        });
        return;
      }

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(team),
      });
    }
  );

  // POST /teams - Create team
  await page.route(`${API_BASE}/teams`, async (route: Route) => {
    if (route.request().method() !== 'POST') return;

    const postData = route.request().postDataJSON();
    const newTeam = {
      id: `team-new-${Date.now()}`,
      org_id: mockOrganization.id,
      name: postData.name,
      description: postData.description || null,
      member_count: 0,
      agent_config_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    await route.fulfill({
      status: 201,
      contentType: 'application/json',
      body: JSON.stringify(newTeam),
    });
  });
}

/**
 * Mock role endpoints
 */
async function mockRoleEndpoints(page: Page): Promise<void> {
  // GET /roles - List all roles
  await page.route(`${API_BASE}/roles`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: mockRoles,
      }),
    });
  });
}

/**
 * Mock agent catalog endpoints
 */
async function mockAgentEndpoints(page: Page): Promise<void> {
  // GET /agents - List available agents
  await page.route(`${API_BASE}/agents`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        agents: mockAgents,
        total: mockAgents.length,
      }),
    });
  });
}

/**
 * Mock agent configuration endpoints
 */
async function mockAgentConfigEndpoints(page: Page): Promise<void> {
  // GET /organizations/current/agent-configs
  await page.route(
    `${API_BASE}/organizations/current/agent-configs`,
    async (route: Route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          configs: mockOrgAgentConfigs,
          total: mockOrgAgentConfigs.length,
        }),
      });
    }
  );

  // POST /organizations/current/agent-configs
  await page.route(
    `${API_BASE}/organizations/current/agent-configs`,
    async (route: Route) => {
      if (route.request().method() !== 'POST') return;

      const postData = route.request().postDataJSON();
      const agent = mockAgents.find((a) => a.id === postData.agent_id);

      const newConfig = {
        id: `org-config-new-${Date.now()}`,
        org_id: mockOrganization.id,
        agent_id: postData.agent_id,
        agent_name: agent?.name || 'Unknown',
        agent_type: agent?.type || 'unknown',
        agent_provider: agent?.provider || 'unknown',
        is_enabled: postData.is_enabled ?? true,
        config: postData.config,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(newConfig),
      });
    }
  );

  // GET /teams/:team_id/agent-configs
  await page.route(
    new RegExp(`${API_BASE}/teams/[a-f0-9-]+/agent-configs$`),
    async (route: Route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          configs: mockTeamAgentConfigs,
          total: mockTeamAgentConfigs.length,
        }),
      });
    }
  );

  // GET /employees/:employee_id/agent-configs
  await page.route(
    new RegExp(`${API_BASE}/employees/[a-f0-9-]+/agent-configs$`),
    async (route: Route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          configs: mockEmployeeAgentConfigs,
          total: mockEmployeeAgentConfigs.length,
        }),
      });
    }
  );

  // GET /employees/:employee_id/agent-configs/resolved
  await page.route(
    new RegExp(`${API_BASE}/employees/[a-f0-9-]+/agent-configs/resolved$`),
    async (route: Route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          configs: mockResolvedAgentConfigs,
          total: mockResolvedAgentConfigs.length,
        }),
      });
    }
  );
}

/**
 * Set up minimal mocks for authentication only
 * Use this for tests that don't need full API mocking
 */
export async function setupAuthMocks(page: Page): Promise<void> {
  await mockAuthEndpoints(page);
}

/**
 * Set up mocks for specific feature areas
 */
export async function setupEmployeeMocks(page: Page): Promise<void> {
  await mockAuthEndpoints(page);
  await mockEmployeeEndpoints(page);
  await mockTeamEndpoints(page);
  await mockRoleEndpoints(page);
}

export async function setupAgentMocks(page: Page): Promise<void> {
  await mockAuthEndpoints(page);
  await mockAgentEndpoints(page);
  await mockAgentConfigEndpoints(page);
}

/**
 * Helper to create a mock user session
 */
export async function mockUserSession(
  page: Page,
  employee = mockEmployees[0]
): Promise<void> {
  // Set auth cookie
  await page.context().addCookies([
    {
      name: 'auth_token',
      value: 'mock-jwt-token-12345',
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax',
    },
  ]);

  // Mock /auth/me to return the employee
  await page.route(`${API_BASE}/auth/me`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(employee),
    });
  });
}
