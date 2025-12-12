import { getCurrentEmployee } from '@/lib/auth';
import { redirect } from 'next/navigation';
import { SettingsSidebar } from '@/components/settings/SettingsSidebar';

export default async function SettingsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const employee = await getCurrentEmployee();

  if (!employee) {
    redirect('/login');
  }

  // Check if user has admin permissions
  const isAdmin = employee.role_name === 'Admin' || employee.role_name === 'admin';

  return (
    <div className="flex gap-8">
      <SettingsSidebar isAdmin={isAdmin} />
      <div className="flex-1 max-w-4xl">{children}</div>
    </div>
  );
}
