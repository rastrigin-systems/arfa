import { getCurrentEmployee, getServerToken } from '@/lib/auth';
import { redirect } from 'next/navigation';
import { DashboardHeader } from '@/components/dashboard-header';
import { Sidebar } from '@/components/layout/Sidebar';
import type { components } from '@/lib/api/schema';

type ListTeamsResponse = components['schemas']['ListTeamsResponse'];
type ListEmployeesResponse = components['schemas']['ListEmployeesResponse'];

async function getStats(token: string): Promise<{ teamCount: number; employeeCount: number }> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

  try {
    const [teamsRes, employeesRes] = await Promise.all([
      fetch(`${apiUrl}/teams`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      }),
      fetch(`${apiUrl}/employees?per_page=1`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      }),
    ]);

    const teams: ListTeamsResponse = teamsRes.ok ? await teamsRes.json() : { teams: [] };
    const employees: ListEmployeesResponse = employeesRes.ok
      ? await employeesRes.json()
      : { employees: [], total: 0, limit: 1, offset: 0 };

    return {
      teamCount: teams.teams?.length ?? 0,
      employeeCount: employees.total ?? 0,
    };
  } catch {
    return { teamCount: 0, employeeCount: 0 };
  }
}

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  // Verify authentication
  const employee = await getCurrentEmployee();

  if (!employee) {
    redirect('/login');
  }

  const token = await getServerToken();
  const stats = token ? await getStats(token) : { teamCount: 0, employeeCount: 0 };

  return (
    <div className="min-h-screen bg-background">
      <DashboardHeader employee={employee} />
      <div className="flex">
        <Sidebar teamCount={stats.teamCount} employeeCount={stats.employeeCount} />
        <main className="ml-64 flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
