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

  // TODO: Implement proper role-based access when API supports role_name
  // For now, show organization settings to all users
  // The sidebar will be visible to everyone
  const isAdmin = true;

  return (
    <div className="flex gap-8">
      <SettingsSidebar isAdmin={isAdmin} />
      <div className="flex-1 max-w-4xl">{children}</div>
    </div>
  );
}
