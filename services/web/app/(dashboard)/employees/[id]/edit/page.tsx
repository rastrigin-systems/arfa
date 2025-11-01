import { redirect } from 'next/navigation';
import { EmployeeForm } from '@/components/employees/EmployeeForm';
import { getServerToken } from '@/lib/auth';
import type { components } from '@/lib/api/schema';

type Employee = components['schemas']['Employee'];

async function getEmployee(id: string, token: string): Promise<Employee> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

  const response = await fetch(`${apiUrl}/employees/${id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    cache: 'no-store', // Always fetch fresh data
  });

  if (!response.ok) {
    if (response.status === 401) {
      redirect('/login');
    }
    if (response.status === 404) {
      redirect('/employees');
    }
    throw new Error(`Failed to fetch employee: ${response.statusText}`);
  }

  return response.json();
}

export default async function EditEmployeePage({
  params,
}: {
  params: { id: string };
}) {
  // Get auth token from cookies
  const token = await getServerToken();

  if (!token) {
    redirect('/login');
  }

  let employee: Employee;
  let error: string | null = null;

  try {
    employee = await getEmployee(params.id, token);
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to load employee';
    // Redirect to employees list on error
    redirect('/employees');
  }

  return (
    <div className="container mx-auto py-8 max-w-2xl space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Edit Employee</h1>
        <p className="text-muted-foreground mt-1">
          Update {employee.full_name}&apos;s information
        </p>
      </div>

      {/* Error Message (if any) */}
      {error && (
        <div className="rounded-md bg-destructive/15 p-4 text-destructive">
          <p className="font-medium">Error loading employee</p>
          <p className="text-sm mt-1">{error}</p>
        </div>
      )}

      {/* Employee Form */}
      {!error && <EmployeeForm mode="edit" employee={employee} />}
    </div>
  );
}
