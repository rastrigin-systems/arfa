import { redirect, notFound } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Settings, UserPlus, Users, Bot } from 'lucide-react';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import type { components } from '@/lib/api/schema';

type Team = components['schemas']['Team'];
type Employee = components['schemas']['Employee'];

async function getTeam(token: string, teamId: string): Promise<Team | null> {
  const { data, error, response } = await apiClient.GET('/teams/{team_id}', {
    params: { path: { team_id: teamId } },
    headers: { Authorization: `Bearer ${token}` },
  });

  if (response.status === 404) {
    return null;
  }

  if (error) {
    if (response.status === 401) {
      redirect('/login');
    }
    throw new Error('Failed to fetch team');
  }

  return data ?? null;
}

async function getTeamMembers(token: string, teamId: string): Promise<Employee[]> {
  const { data } = await apiClient.GET('/employees', {
    params: { query: { per_page: 100 } },
    headers: { Authorization: `Bearer ${token}` },
  });

  // Filter employees by team_id
  return (data?.employees ?? []).filter((emp) => emp.team_id === teamId);
}

function getStatusVariant(status: string) {
  switch (status) {
    case 'active':
      return 'success';
    case 'suspended':
      return 'warning';
    case 'inactive':
      return 'secondary';
    default:
      return 'default';
  }
}

interface TeamDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function TeamDetailPage({ params }: TeamDetailPageProps) {
  const { id: teamId } = await params;
  const token = await getServerToken();

  if (!token) {
    redirect('/login');
  }

  const [team, members] = await Promise.all([
    getTeam(token, teamId),
    getTeamMembers(token, teamId),
  ]);

  if (!team) {
    notFound();
  }

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link href="/teams" className="hover:text-foreground transition-colors">
          Teams
        </Link>
        <span>/</span>
        <span className="text-foreground">{team.name}</span>
      </div>

      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="/teams">
            <Button variant="ghost" size="icon">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{team.name}</h1>
            {team.description && (
              <p className="text-muted-foreground mt-1">{team.description}</p>
            )}
          </div>
        </div>
        <div className="flex gap-2">
          <Link href={`/teams/${teamId}/edit`}>
            <Button variant="outline" className="gap-2">
              <Settings className="h-4 w-4" />
              Edit Team
            </Button>
          </Link>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Team Members</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{members.length}</div>
            <p className="text-xs text-muted-foreground">
              Active employees in this team
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Agent Configs</CardTitle>
            <Bot className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">-</div>
            <p className="text-xs text-muted-foreground">
              Team-level agent configurations
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Created</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {team.created_at
                ? new Date(team.created_at).toLocaleDateString()
                : '-'}
            </div>
            <p className="text-xs text-muted-foreground">
              Team creation date
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Team Members */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Team Members</CardTitle>
              <CardDescription>Employees assigned to this team</CardDescription>
            </div>
            <Link href={`/employees/new?team=${teamId}`}>
              <Button size="sm" className="gap-2">
                <UserPlus className="h-4 w-4" />
                Add Member
              </Button>
            </Link>
          </div>
        </CardHeader>
        <CardContent>
          {members.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <Users className="h-12 w-12 text-muted-foreground/50" />
              <h3 className="mt-4 text-lg font-semibold">No members yet</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                Add employees to this team to get started.
              </p>
              <Link href={`/employees/new?team=${teamId}`} className="mt-4">
                <Button size="sm" className="gap-2">
                  <UserPlus className="h-4 w-4" />
                  Add first member
                </Button>
              </Link>
            </div>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {members.map((member) => (
                    <TableRow key={member.id}>
                      <TableCell className="font-medium">{member.full_name}</TableCell>
                      <TableCell>{member.email}</TableCell>
                      <TableCell>
                        <Badge variant={getStatusVariant(member.status)}>
                          {member.status}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Link href={`/employees/${member.id}`}>
                          <Button variant="ghost" size="sm">
                            View
                          </Button>
                        </Link>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
