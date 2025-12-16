import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { getCurrentEmployee } from '@/lib/auth';
import { Bot, Users, Settings } from 'lucide-react';

export default async function DashboardPage() {
  const employee = await getCurrentEmployee();

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
        <p className="text-muted-foreground">
          Welcome back, {employee?.full_name}!
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bot className="h-5 w-5 text-muted-foreground" />
              <CardTitle>Agent Configurations</CardTitle>
            </div>
            <CardDescription>Manage your AI agent settings</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/agents">
              <Button variant="outline" size="sm">View Agents</Button>
            </Link>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Users className="h-5 w-5 text-muted-foreground" />
              <CardTitle>Team Management</CardTitle>
            </div>
            <CardDescription>Manage teams and members</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/teams">
              <Button variant="outline" size="sm">View Teams</Button>
            </Link>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Settings className="h-5 w-5 text-muted-foreground" />
              <CardTitle>Organization Settings</CardTitle>
            </div>
            <CardDescription>Configure organization preferences</CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/settings">
              <Button variant="outline" size="sm">View Settings</Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
