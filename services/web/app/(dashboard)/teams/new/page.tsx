import { redirect } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { getServerToken } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { TeamForm } from '@/components/teams/TeamForm';

export default async function NewTeamPage() {
  const token = await getServerToken();

  if (!token) {
    redirect('/login');
  }

  return (
    <div className="space-y-6 max-w-2xl">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link href="/teams" className="hover:text-foreground transition-colors">
          Teams
        </Link>
        <span>/</span>
        <span className="text-foreground">New Team</span>
      </div>

      {/* Page Header */}
      <div className="flex items-center gap-4">
        <Link href="/teams">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Create Team</h1>
          <p className="text-muted-foreground mt-1">
            Create a new team to organize employees
          </p>
        </div>
      </div>

      {/* Form */}
      <TeamForm token={token} />
    </div>
  );
}
