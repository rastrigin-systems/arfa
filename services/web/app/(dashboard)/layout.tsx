import { getCurrentEmployee } from '@/lib/auth';
import { redirect } from 'next/navigation';
import { DashboardHeader } from '@/components/dashboard-header';

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

  return (
    <div className="min-h-screen bg-background">
      <DashboardHeader employee={employee} />
      <main className="container mx-auto p-6">{children}</main>
    </div>
  );
}
