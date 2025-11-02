'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { EmployeeAgentOverridesTable, type EmployeeAgentOverride } from '@/components/agents/EmployeeAgentOverridesTable';
import { ArrowLeft } from 'lucide-react';

type Employee = {
  readonly id: string;
  org_id: string;
  team_id?: string | null;
  readonly team_name?: string | null;
  role_id: string;
  email: string;
  full_name: string;
  status: 'active' | 'suspended' | 'inactive';
  preferences?: Record<string, unknown>;
  readonly created_at?: string;
  readonly updated_at?: string;
};

type EmployeeAgentOverridesClientProps = {
  employee: Employee;
  agentOverrides: EmployeeAgentOverride[];
};

export function EmployeeAgentOverridesClient({
  employee,
  agentOverrides,
}: EmployeeAgentOverridesClientProps) {
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedOverride, setSelectedOverride] = useState<EmployeeAgentOverride | null>(null);

  const handleEdit = (override: EmployeeAgentOverride) => {
    // TODO: Implement edit modal
    console.log('Edit override:', override.id);
    alert('Edit functionality coming soon');
  };

  const handleDelete = (override: EmployeeAgentOverride) => {
    setSelectedOverride(override);
    setShowDeleteModal(true);
  };

  const handleToggleEnabled = async (override: EmployeeAgentOverride) => {
    // TODO: Implement toggle enabled API call
    console.log('Toggle enabled:', override.id);
    alert('Toggle functionality coming soon');
  };

  const handleConfirmDelete = async () => {
    // TODO: Implement delete API call
    console.log('Delete override:', selectedOverride?.id);
    alert('Delete functionality coming soon');
    setShowDeleteModal(false);
    setSelectedOverride(null);
  };

  return (
    <div className="space-y-6">
      {/* Header with breadcrumb */}
      <div>
        <div className="flex items-center gap-2 mb-2">
          <Link
            href="/employees"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            Employees
          </Link>
          <span className="text-sm text-muted-foreground">/</span>
          <Link
            href={`/employees/${employee.id}`}
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            {employee.full_name}
          </Link>
          <span className="text-sm text-muted-foreground">/</span>
          <span className="text-sm font-medium">Agent Overrides</span>
        </div>
        <h1 className="text-3xl font-bold">{employee.full_name} - Agent Overrides</h1>
        <p className="text-muted-foreground mt-2">
          Manage agent configurations specific to this employee
        </p>
      </div>

      {/* Back button */}
      <div>
        <Link href={`/employees/${employee.id}`}>
          <Button variant="outline" size="sm">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Employee
          </Button>
        </Link>
      </div>

      {/* Employee Agent Overrides Card */}
      <Card>
        <CardHeader>
          <CardTitle>Employee Agent Overrides</CardTitle>
          <CardDescription>
            Employee-specific agent configuration overrides
          </CardDescription>
        </CardHeader>
        <CardContent>
          <EmployeeAgentOverridesTable
            overrides={agentOverrides}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onToggleEnabled={handleToggleEnabled}
          />
        </CardContent>
      </Card>

      {/* Delete Confirmation Modal */}
      {showDeleteModal && selectedOverride && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
          onClick={() => {
            setShowDeleteModal(false);
            setSelectedOverride(null);
          }}
        >
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="delete-modal-title"
            className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 id="delete-modal-title" className="text-xl font-bold mb-2">
              Remove Override
            </h2>
            <p className="text-sm text-muted-foreground mb-4">
              Remove <strong>{selectedOverride.agent_name}</strong> override for{' '}
              <strong>{employee.full_name}</strong>?
            </p>
            <p className="text-sm text-muted-foreground mb-6">
              Employee will use default organization/team configuration.
            </p>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setShowDeleteModal(false);
                  setSelectedOverride(null);
                }}
              >
                Cancel
              </Button>
              <Button variant="destructive" onClick={handleConfirmDelete}>
                Remove Override
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
