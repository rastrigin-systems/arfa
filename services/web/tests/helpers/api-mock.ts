/**
 * API Mocking Helpers
 *
 * Utilities for mocking API responses in E2E tests
 */

import { Page, Route } from '@playwright/test';
import {
  mockEmployees,
  mockEmployee,
  mockCreateEmployeeResponse,
  mockAgents,
  mockAgent,
  mockAgentConfigs,
} from '../fixtures';

const API_BASE_URL = 'http://localhost:8080/api/v1';

/**
 * Mock employee list API endpoint
 *
 * @param page - Playwright Page object
 * @param employees - Employee data to return (defaults to mockEmployees)
 */
export async function mockGetEmployees(
  page: Page,
  employees: any[] = mockEmployees
) {
  await page.route(`${API_BASE_URL}/employees*`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ employees, total: employees.length }),
    });
  });
}

/**
 * Mock single employee API endpoint
 *
 * @param page - Playwright Page object
 * @param employee - Employee data to return (defaults to mockEmployee)
 */
export async function mockGetEmployee(
  page: Page,
  employeeId: string,
  employee: any = mockEmployee
) {
  await page.route(`${API_BASE_URL}/employees/${employeeId}`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(employee),
    });
  });
}

/**
 * Mock employee creation API endpoint
 *
 * @param page - Playwright Page object
 * @param response - Employee creation response (defaults to mockCreateEmployeeResponse)
 */
export async function mockCreateEmployee(
  page: Page,
  response: any = mockCreateEmployeeResponse
) {
  await page.route(`${API_BASE_URL}/employees`, async (route: Route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(response),
      });
    } else {
      await route.continue();
    }
  });
}

/**
 * Mock employee update API endpoint
 *
 * @param page - Playwright Page object
 * @param employeeId - Employee ID to update
 * @param response - Updated employee data
 */
export async function mockUpdateEmployee(
  page: Page,
  employeeId: string,
  response: any = mockEmployee
) {
  await page.route(`${API_BASE_URL}/employees/${employeeId}`, async (route: Route) => {
    if (route.request().method() === 'PUT' || route.request().method() === 'PATCH') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(response),
      });
    } else {
      await route.continue();
    }
  });
}

/**
 * Mock employee deletion API endpoint
 *
 * @param page - Playwright Page object
 * @param employeeId - Employee ID to delete
 */
export async function mockDeleteEmployee(page: Page, employeeId: string) {
  await page.route(`${API_BASE_URL}/employees/${employeeId}`, async (route: Route) => {
    if (route.request().method() === 'DELETE') {
      await route.fulfill({
        status: 204,
      });
    } else {
      await route.continue();
    }
  });
}

/**
 * Mock agent catalog API endpoint
 *
 * @param page - Playwright Page object
 * @param agents - Agent data to return (defaults to mockAgents)
 */
export async function mockGetAgents(page: Page, agents: any[] = mockAgents) {
  await page.route(`${API_BASE_URL}/agents*`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ agents, total: agents.length }),
    });
  });
}

/**
 * Mock organization agent configs API endpoint
 *
 * @param page - Playwright Page object
 * @param configs - Agent config data to return (defaults to mockAgentConfigs)
 */
export async function mockGetOrgAgentConfigs(
  page: Page,
  configs: any[] = mockAgentConfigs
) {
  await page.route(`${API_BASE_URL}/organizations/*/agent-configs*`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ configs, total: configs.length }),
    });
  });
}

/**
 * Mock API error response
 *
 * @param page - Playwright Page object
 * @param url - URL pattern to match
 * @param status - HTTP status code
 * @param message - Error message
 */
export async function mockAPIError(
  page: Page,
  url: string,
  status: number = 500,
  message: string = 'Internal Server Error'
) {
  await page.route(url, async (route: Route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ error: message }),
    });
  });
}

/**
 * Clear all API route mocks
 *
 * @param page - Playwright Page object
 */
export async function clearAPIMocks(page: Page) {
  await page.unrouteAll({ behavior: 'ignoreErrors' });
}
