'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import Link from 'next/link';
import type { Employee } from '@/lib/types';

type EmployeeDetailClientProps = {
  employee: Employee;
};

function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' {
  switch (status) {
    case 'active':
      return 'default';
    case 'inactive':
      return 'secondary';
    case 'suspended':
      return 'destructive';
    default:
      return 'outline' as 'default';
  }
}

export function EmployeeDetailClient({ employee }: EmployeeDetailClientProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const handleEdit = () => {
    router.push(`/employees/${employee.id}/edit`);
  };

  const handleDelete = () => {
    setShowDeleteModal(true);
  };

  const handleConfirmDelete = async () => {
    // TODO: Implement delete API call when backend supports it
    toast({
      title: 'Not implemented',
      description: 'Delete functionality is not yet available',
      variant: 'destructive',
    });
    setShowDeleteModal(false);
  };

  const formatDate = (dateString?: string | null) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Link
              href="/employees"
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Employees
            </Link>
            <span className="text-sm text-muted-foreground">/</span>
            <span className="text-sm font-medium">{employee.full_name}</span>
          </div>
          <h1 className="text-3xl font-bold">Employee Details</h1>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleEdit}>
            Edit
          </Button>
          <Button variant="destructive" onClick={handleDelete}>
            Delete
          </Button>
        </div>
      </div>

      {/* Basic Information Card */}
      <Card>
        <CardHeader>
          <CardTitle>Basic Information</CardTitle>
          <CardDescription>Employee profile and status</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Name</dt>
              <dd className="mt-1 text-sm">{employee.full_name}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Email</dt>
              <dd className="mt-1 text-sm">{employee.email}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Status</dt>
              <dd className="mt-1">
                <Badge
                  variant={getStatusVariant(employee.status)}
                  data-testid="employee-status-badge"
                >
                  {employee.status}
                </Badge>
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Team</dt>
              <dd className="mt-1 text-sm">{employee.team_name || 'No team assigned'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Last Login</dt>
              <dd className="mt-1 text-sm">{formatDate(employee.last_login_at)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">Personal Claude Token</dt>
              <dd className="mt-1 text-sm">
                {employee.has_personal_claude_token ? 'Configured' : 'Not configured'}
              </dd>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Teams Card */}
      <Card>
        <CardHeader>
          <CardTitle>Teams</CardTitle>
          <CardDescription>Team memberships for this employee</CardDescription>
        </CardHeader>
        <CardContent>
          <div data-testid="employee-teams">
            {employee.team_name ? (
              <div className="flex items-center gap-2">
                <Badge variant="outline">{employee.team_name}</Badge>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No teams assigned</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* MCP Servers Card */}
      <Card>
        <CardHeader>
          <CardTitle>MCP Servers</CardTitle>
          <CardDescription>Model Context Protocol servers for this employee</CardDescription>
        </CardHeader>
        <CardContent>
          <div data-testid="employee-mcps">
            <p className="text-sm text-muted-foreground">No MCP servers configured</p>
            {/* TODO: Fetch and display MCP configs when API is available */}
          </div>
        </CardContent>
      </Card>

      {/* Delete Confirmation Modal */}
      {showDeleteModal && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
          onClick={() => setShowDeleteModal(false)}
        >
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="delete-modal-title"
            className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 id="delete-modal-title" className="text-xl font-bold mb-2">
              Delete Employee
            </h2>
            <p className="text-sm text-muted-foreground mb-6">
              Are you sure you want to delete <strong>{employee.full_name}</strong>? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setShowDeleteModal(false)}>
                Cancel
              </Button>
              <Button variant="destructive" onClick={handleConfirmDelete}>
                Confirm Delete
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
