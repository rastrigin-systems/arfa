import { redirect } from 'next/navigation';
import Link from 'next/link';
import { Plus, UsersRound } from 'lucide-react';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { TeamCard } from '@/components/teams/TeamCard';

export default async function TeamsPage() {
  const token = await getServerToken();

  if (!token) {
    redirect('/login');
  }

  const { data: teams, error: apiError } = await apiClient.GET('/teams', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  const error = apiError ? 'Failed to load teams' : null;
  const teamList = teams?.teams ?? [];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Teams</h1>
          <p className="text-muted-foreground mt-1">
            Organize employees into teams for better management
          </p>
        </div>
        <Link href="/teams/new">
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            Create Team
          </Button>
        </Link>
      </div>

      {/* Error Message */}
      {error && (
        <div className="rounded-md bg-destructive/15 p-4 text-destructive">
          <p className="font-medium">Error loading teams</p>
          <p className="text-sm mt-1">{error}</p>
        </div>
      )}

      {/* Teams Grid */}
      {teamList.length === 0 && !error ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
            <UsersRound className="h-6 w-6 text-muted-foreground" />
          </div>
          <h3 className="mt-4 text-lg font-semibold">No teams yet</h3>
          <p className="mt-2 text-sm text-muted-foreground max-w-sm">
            Create your first team to start organizing employees and managing agent configurations.
          </p>
          <Link href="/teams/new" className="mt-6">
            <Button className="gap-2">
              <Plus className="h-4 w-4" />
              Create your first team
            </Button>
          </Link>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {teamList.map((team) => (
            <TeamCard key={team.id} team={team} />
          ))}
        </div>
      )}
    </div>
  );
}
