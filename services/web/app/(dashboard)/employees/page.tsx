import { redirect } from 'next/navigation';
import Link from 'next/link';
import { Plus, Users } from 'lucide-react';
import { EmployeeTable } from '@/components/employees/EmployeeTable';
import { Button } from '@/components/ui/button';
import { getServerToken } from '@/lib/auth';
import type { components } from '@/lib/api/schema';

type ListEmployeesResponse = components['schemas']['ListEmployeesResponse'];
type ListTeamsResponse = components['schemas']['ListTeamsResponse'];

async function getEmployees(token: string): Promise<ListEmployeesResponse> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

  const response = await fetch(`${apiUrl}/employees`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    cache: 'no-store',
  });

  if (!response.ok) {
    if (response.status === 401) {
      redirect('/login');
    }
    throw new Error(`Failed to fetch employees: ${response.statusText}`);
  }

  return response.json();
}

async function getTeams(token: string): Promise<ListTeamsResponse> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

  const response = await fetch(`${apiUrl}/teams`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    cache: 'no-store',
  });

  if (!response.ok) {
    return { teams: [], total: 0 };
  }

  return response.json();
}

export default async function EmployeesPage() {
  const token = await getServerToken();

  if (!token) {
    redirect('/login');
  }

  let employees: ListEmployeesResponse;
  let teams: ListTeamsResponse;
  let error: string | null = null;

  try {
    [employees, teams] = await Promise.all([getEmployees(token), getTeams(token)]);
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to load employees';
    employees = {
      employees: [],
      total: 0,
      limit: 20,
      offset: 0,
    };
    teams = { teams: [], total: 0 };
  }

  const employeeList = employees.employees ?? [];
  const teamList = teams.teams ?? [];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Employees</h1>
          <p className="text-muted-foreground mt-1">
            Manage employees in your organization
          </p>
        </div>
        <Link href="/employees/new">
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            Add Employee
          </Button>
        </Link>
      </div>

      {/* Error Message */}
      {error && (
        <div className="rounded-md bg-destructive/15 p-4 text-destructive">
          <p className="font-medium">Error loading employees</p>
          <p className="text-sm mt-1">{error}</p>
        </div>
      )}

      {/* Empty State */}
      {employeeList.length === 0 && !error ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
            <Users className="h-6 w-6 text-muted-foreground" />
          </div>
          <h3 className="mt-4 text-lg font-semibold">No employees yet</h3>
          <p className="mt-2 text-sm text-muted-foreground max-w-sm">
            Add your first employee to start managing your organization.
          </p>
          <Link href="/employees/new" className="mt-6">
            <Button className="gap-2">
              <Plus className="h-4 w-4" />
              Add your first employee
            </Button>
          </Link>
        </div>
      ) : (
        <EmployeeTable
          employees={employeeList}
          teams={teamList}
          total={employees.total}
        />
      )}
    </div>
  );
}
