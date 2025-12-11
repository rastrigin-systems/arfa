import { redirect } from 'next/navigation';
import Link from 'next/link';
import { Plus, Users } from 'lucide-react';
import { EmployeeTable } from '@/components/employees/EmployeeTable';
import { Button } from '@/components/ui/button';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';

export default async function EmployeesPage() {
  const token = await getServerToken();

  if (!token) {
    redirect('/login');
  }

  const [employeesRes, teamsRes] = await Promise.all([
    apiClient.GET('/employees', {
      headers: { Authorization: `Bearer ${token}` },
    }),
    apiClient.GET('/teams', {
      headers: { Authorization: `Bearer ${token}` },
    }),
  ]);

  const error = employeesRes.error ? 'Failed to load employees' : null;
  const employeeList = employeesRes.data?.employees ?? [];
  const teamList = teamsRes.data?.teams ?? [];
  const total = employeesRes.data?.total ?? 0;

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
          total={total}
        />
      )}
    </div>
  );
}
