'use client';

import Link from 'next/link';
import { UsersRound, Settings, ChevronRight } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { components } from '@/lib/api/schema';

type Team = components['schemas']['Team'];

interface TeamCardProps {
  team: Team;
}

export function TeamCard({ team }: TeamCardProps) {
  const memberCount = team.member_count ?? 0;
  const configCount = team.agent_config_count ?? 0;

  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
              <UsersRound className="h-5 w-5 text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg">{team.name}</CardTitle>
              {team.description && (
                <CardDescription className="mt-1 line-clamp-2">
                  {team.description}
                </CardDescription>
              )}
            </div>
          </div>
          <Link href={`/teams/${team.id}/edit`}>
            <Button variant="ghost" size="icon" className="h-8 w-8">
              <Settings className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <div className="flex gap-4">
            <Badge variant="secondary" className="gap-1">
              <UsersRound className="h-3 w-3" />
              {memberCount} {memberCount === 1 ? 'member' : 'members'}
            </Badge>
            {configCount > 0 && (
              <Badge variant="outline" className="gap-1">
                {configCount} agent {configCount === 1 ? 'config' : 'configs'}
              </Badge>
            )}
          </div>
          <Link href={`/teams/${team.id}`}>
            <Button variant="ghost" size="sm" className="gap-1">
              View
              <ChevronRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}
