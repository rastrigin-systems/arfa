import { redirect, notFound } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { getServerToken } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { TeamForm } from '@/components/teams/TeamForm';
import type { components } from '@/lib/api/schema';

type Team = components['schemas']['Team'];

async function getTeam(token: string, teamId: string): Promise<Team | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

  const response = await fetch(`${apiUrl}/teams/${teamId}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    cache: 'no-store',
  });

  if (response.status === 404) {
    return null;
  }

  if (!response.ok) {
    if (response.status === 401) {
      redirect('/login');
    }
    throw new Error(`Failed to fetch team: ${response.statusText}`);
  }

  return response.json();
}

interface EditTeamPageProps {
  params: Promise<{ id: string }>;
}

export default async function EditTeamPage({ params }: EditTeamPageProps) {
  const { id: teamId } = await params;
  const token = await getServerToken();

  if (!token) {
    redirect('/login');
  }

  const team = await getTeam(token, teamId);

  if (!team) {
    notFound();
  }

  return (
    <div className="space-y-6 max-w-2xl">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link href="/teams" className="hover:text-foreground transition-colors">
          Teams
        </Link>
        <span>/</span>
        <Link href={`/teams/${teamId}`} className="hover:text-foreground transition-colors">
          {team.name}
        </Link>
        <span>/</span>
        <span className="text-foreground">Edit</span>
      </div>

      {/* Page Header */}
      <div className="flex items-center gap-4">
        <Link href={`/teams/${teamId}`}>
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Edit Team</h1>
          <p className="text-muted-foreground mt-1">
            Update team details for {team.name}
          </p>
        </div>
      </div>

      {/* Form */}
      <TeamForm team={team} token={token} />
    </div>
  );
}
