import { getCurrentEmployee } from '@/lib/auth';
import { redirect } from 'next/navigation';
import { ProfileSettingsClient } from './ProfileSettingsClient';

export default async function ProfileSettingsPage() {
  const employee = await getCurrentEmployee();

  if (!employee) {
    redirect('/login');
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <nav className="text-sm text-muted-foreground mb-2">
          Settings &gt; Profile
        </nav>
        <h1 className="text-3xl font-bold tracking-tight">Profile & Preferences</h1>
        <p className="text-muted-foreground mt-1">
          Manage your personal settings and preferences
        </p>
      </div>

      <ProfileSettingsClient employee={employee} />
    </div>
  );
}
