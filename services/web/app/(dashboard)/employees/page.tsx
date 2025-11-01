import { redirect } from 'next/navigation';
import { EmployeeTable } from '@/components/employees/EmployeeTable';
import { Button } from '@/components/ui/button';
import { getServerToken } from '@/lib/auth';
import type { components } from '@/lib/api/schema';

type ListEmployeesResponse = components['schemas']['ListEmployeesResponse'];

async function getEmployees(token: string): Promise<ListEmployeesResponse> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

  const response = await fetch(`${apiUrl}/employees`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    cache: 'no-store', // Always fetch fresh data
  });

  if (!response.ok) {
    if (response.status === 401) {
      redirect('/login');
    }
    throw new Error(`Failed to fetch employees: ${response.statusText}`);
  }

  return response.json();
}

export default async function EmployeesPage() {
  // Get auth token from cookies
  const token = await getServerToken();

  if (!token) {
    redirect('/login');
  }

  let employees: ListEmployeesResponse;
  let error: string | null = null;

  try {
    employees = await getEmployees(token);
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to load employees';
    // For now, show empty state on error
    employees = {
      employees: [],
      total: 0,
      limit: 20,
      offset: 0,
    };
  }

  return (
    <div className="container mx-auto py-8 space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Employees</h1>
          <p className="text-muted-foreground mt-1">
            Manage employees in your organization
          </p>
        </div>
        <Button onClick={() => { /* TODO: Navigate to create page */ }}>
          Create Employee
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <div className="rounded-md bg-destructive/15 p-4 text-destructive">
          <p className="font-medium">Error loading employees</p>
          <p className="text-sm mt-1">{error}</p>
        </div>
      )}

      {/* Employee Table */}
      <EmployeeTable
        employees={employees.employees}
        total={employees.total}
      />
    </div>
  );
}
