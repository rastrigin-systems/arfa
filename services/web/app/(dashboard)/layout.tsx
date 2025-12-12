import { getCurrentEmployee, getServerToken } from '@/lib/auth';
import { redirect } from 'next/navigation';
import { DashboardHeader } from '@/components/dashboard-header';
import { Sidebar } from '@/components/layout/Sidebar';
import { apiClient } from '@/lib/api/client';

async function getStats(token: string): Promise<{ teamCount: number; employeeCount: number }> {
  try {
    const [teamsRes, employeesRes] = await Promise.all([
      apiClient.GET('/teams', {
        headers: { Authorization: `Bearer ${token}` },
      }),
      apiClient.GET('/employees', {
        params: { query: { per_page: 1 } },
        headers: { Authorization: `Bearer ${token}` },
      }),
    ]);

    return {
      teamCount: teamsRes.data?.teams?.length ?? 0,
      employeeCount: employeesRes.data?.total ?? 0,
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
