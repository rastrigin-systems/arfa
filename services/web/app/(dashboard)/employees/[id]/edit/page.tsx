import { redirect } from 'next/navigation';
import { EmployeeForm } from '@/components/employees/EmployeeForm';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import type { components } from '@/lib/api/schema';

type Employee = components['schemas']['Employee'];

async function getEmployee(id: string, token: string): Promise<Employee | null> {
  const { data, error, response } = await apiClient.GET('/employees/{employee_id}', {
    params: { path: { employee_id: id } },
    headers: { Authorization: `Bearer ${token}` },
  });

  if (response.status === 401) {
    redirect('/login');
  }

  if (response.status === 404 || error) {
    return null;
  }

  return data ?? null;
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

  const employee = await getEmployee(params.id, token);

  if (!employee) {
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

      {/* Employee Form */}
      <EmployeeForm mode="edit" employee={employee} />
    </div>
  );
}
