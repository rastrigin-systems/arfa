import { getCurrentEmployee } from '@/lib/auth';
import { redirect } from 'next/navigation';

export default async function SettingsPage() {
  const employee = await getCurrentEmployee();

  if (!employee) {
    redirect('/login');
  }

  // TODO: Add role-based redirect when API supports role_name
  // For now, redirect all users to profile settings
  // Admin users can access organization settings via the sidebar
  redirect('/settings/profile');
}
