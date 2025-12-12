import { getCurrentEmployee } from '@/lib/auth';
import { redirect } from 'next/navigation';

export default async function SettingsPage() {
  const employee = await getCurrentEmployee();

  if (!employee) {
    redirect('/login');
  }

  // Check if user has admin permissions
  const isAdmin = employee.role_name === 'Admin' || employee.role_name === 'admin';

  // Redirect based on role
  if (isAdmin) {
    redirect('/settings/organization');
  } else {
    redirect('/settings/profile');
  }
}
