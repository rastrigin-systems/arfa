'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ChevronDown, ChevronRight, Building2, Users, User, CheckCircle2, ArrowUp } from 'lucide-react';

type OrgAgentConfig = {
  readonly id: string;
  org_id: string;
  agent_id: string;
  readonly agent_name?: string;
  readonly agent_type?: string;
  readonly agent_provider?: string;
  config: Record<string, unknown>;
  is_enabled: boolean;
  readonly created_at?: string;
  readonly updated_at?: string;
};

type TeamAgentConfig = {
  readonly id: string;
  team_id: string;
  agent_id: string;
  readonly agent_name?: string;
  readonly agent_type?: string;
  readonly agent_provider?: string;
  config_override: Record<string, unknown>;
  is_enabled: boolean;
  readonly created_at?: string;
  readonly updated_at?: string;
};

type EmployeeAgentConfig = {
  readonly id: string;
  employee_id: string;
  agent_id: string;
  readonly agent_name?: string;
  readonly agent_type?: string;
  readonly agent_provider?: string;
  config_override: Record<string, unknown>;
  override_reason?: string | null;
  is_enabled: boolean;
  sync_token?: string | null;
  last_synced_at?: string | null;
  readonly created_at?: string;
  readonly updated_at?: string;
};

type ResolvedAgentConfig = {
  agent_id: string;
  agent_name: string;
  agent_type: string;
  provider: string;
  config: Record<string, unknown>;
  system_prompt?: string;
  is_enabled: boolean;
  sync_token?: string | null;
  last_synced_at?: string | null;
};

type ConfigurationHierarchyProps = {
  orgConfigs: OrgAgentConfig[];
  teamConfigs: TeamAgentConfig[];
  employeeConfigs: EmployeeAgentConfig[];
  resolvedConfigs: ResolvedAgentConfig[];
  teamName?: string | null;
  employeeName: string;
};

function getConfigPreview(config: Record<string, unknown>, maxKeys = 3): string {
  const entries = Object.entries(config).slice(0, maxKeys);
  return entries.map(([key, value]) => {
    const displayValue = typeof value === 'object' ? JSON.stringify(value) : String(value);
    const truncatedValue = displayValue.length > 30 ? displayValue.slice(0, 30) + '...' : displayValue;
    return `${key}: ${truncatedValue}`;
  }).join(', ');
}

function formatDate(dateString?: string | null): string {
  if (!dateString) return 'Never';
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

type ConfigSectionProps = {
  title: string;
  subtitle: string;
  count: number;
  icon: React.ReactNode;
  defaultOpen?: boolean;
  children: React.ReactNode;
};

function ConfigSection({ title, subtitle, count, icon, defaultOpen = true, children }: ConfigSectionProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <div className="border rounded-lg">
      <button
        className="w-full flex items-center justify-between p-4 bg-muted/50 hover:bg-muted transition-colors rounded-t-lg"
        aria-expanded={isOpen}
        onClick={() => setIsOpen(!isOpen)}
      >
        <div className="flex items-center gap-3">
          {icon}
          <div className="text-left">
            <h3 className="font-medium">{title}</h3>
            <p className="text-sm text-muted-foreground">{subtitle}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="secondary">{count} configured</Badge>
          {isOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
        </div>
      </button>
      {isOpen && (
        <div className="p-4 space-y-3">
          {children}
        </div>
      )}
    </div>
  );
}

type ViewConfigModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  agentName: string;
  level: string;
  config: Record<string, unknown>;
  isEnabled: boolean;
  updatedAt?: string | null;
};

function ViewConfigModal({ open, onOpenChange, agentName, level, config, isEnabled, updatedAt }: ViewConfigModalProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{agentName} Configuration ({level})</DialogTitle>
          <DialogDescription>
            Full configuration details for this agent at the {level.toLowerCase()} level
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-2 mb-4">
          <span className="text-sm font-medium">Status:</span>
          <Badge variant={isEnabled ? 'default' : 'secondary'}>
            {isEnabled ? 'Enabled' : 'Disabled'}
          </Badge>
        </div>

        <div className="flex-1 min-h-0 overflow-auto">
          <div className="bg-muted rounded-md p-4">
            <pre className="text-sm font-mono whitespace-pre-wrap overflow-auto">
              {JSON.stringify(config, null, 2)}
            </pre>
          </div>
        </div>

        {updatedAt && (
          <p className="text-sm text-muted-foreground mt-4">
            Last updated: {formatDate(updatedAt)}
          </p>
        )}

        <div className="flex justify-end mt-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

type OrgConfigCardProps = {
  config: OrgAgentConfig;
};

function OrgConfigCard({ config }: OrgConfigCardProps) {
  const [showModal, setShowModal] = useState(false);

  return (
    <>
      <div className="flex items-start justify-between p-4 border rounded-lg bg-card">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <h4 className="font-medium">{config.agent_name || 'Unknown Agent'}</h4>
            <Badge variant={config.is_enabled ? 'default' : 'secondary'} className="text-xs">
              {config.is_enabled ? 'Enabled' : 'Disabled'}
            </Badge>
          </div>
          <p className="text-sm font-mono text-muted-foreground">
            {getConfigPreview(config.config)}
          </p>
        </div>
        <Button variant="ghost" size="sm" onClick={() => setShowModal(true)}>
          View Full Config
        </Button>
      </div>

      <ViewConfigModal
        open={showModal}
        onOpenChange={setShowModal}
        agentName={config.agent_name || 'Unknown Agent'}
        level="Organization"
        config={config.config}
        isEnabled={config.is_enabled}
        updatedAt={config.updated_at}
      />
    </>
  );
}

type TeamConfigCardProps = {
  config: TeamAgentConfig;
};

function TeamConfigCard({ config }: TeamConfigCardProps) {
  const [showModal, setShowModal] = useState(false);
  const overrideKeys = Object.keys(config.config_override);

  return (
    <>
      <div className="flex items-start justify-between p-4 border rounded-lg bg-card">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <h4 className="font-medium">{config.agent_name || 'Unknown Agent'}</h4>
            <Badge variant={config.is_enabled ? 'default' : 'secondary'} className="text-xs">
              {config.is_enabled ? 'Enabled' : 'Disabled'}
            </Badge>
          </div>
          <div className="flex items-center gap-1 text-sm text-muted-foreground">
            <span className="font-medium">Overrides:</span>
            <span className="font-mono">{getConfigPreview(config.config_override)}</span>
            {overrideKeys.length > 0 && (
              <ArrowUp className="h-3 w-3 text-blue-500 ml-1" aria-label="Overrides organization value" />
            )}
          </div>
        </div>
        <Button variant="ghost" size="sm" onClick={() => setShowModal(true)}>
          View Full Config
        </Button>
      </div>

      <ViewConfigModal
        open={showModal}
        onOpenChange={setShowModal}
        agentName={config.agent_name || 'Unknown Agent'}
        level="Team"
        config={config.config_override}
        isEnabled={config.is_enabled}
        updatedAt={config.updated_at}
      />
    </>
  );
}

type EmployeeConfigCardProps = {
  config: EmployeeAgentConfig;
};

function EmployeeConfigCard({ config }: EmployeeConfigCardProps) {
  const [showModal, setShowModal] = useState(false);
  const overrideKeys = Object.keys(config.config_override);

  return (
    <>
      <div className="flex items-start justify-between p-4 border rounded-lg bg-card">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <h4 className="font-medium">{config.agent_name || 'Unknown Agent'}</h4>
            <Badge variant={config.is_enabled ? 'default' : 'secondary'} className="text-xs">
              {config.is_enabled ? 'Enabled' : 'Disabled'}
            </Badge>
          </div>
          <div className="flex items-center gap-1 text-sm text-muted-foreground mb-1">
            <span className="font-medium">Overrides:</span>
            <span className="font-mono">{getConfigPreview(config.config_override)}</span>
            {overrideKeys.length > 0 && (
              <ArrowUp className="h-3 w-3 text-purple-500 ml-1" aria-label="Overrides team/org value" />
            )}
          </div>
          {config.override_reason && (
            <p className="text-sm text-muted-foreground">
              Reason: &quot;{config.override_reason}&quot;
            </p>
          )}
          {config.last_synced_at && (
            <p className="text-xs text-muted-foreground mt-1">
              Last synced: {formatDate(config.last_synced_at)}
            </p>
          )}
        </div>
        <Button variant="ghost" size="sm" onClick={() => setShowModal(true)}>
          View
        </Button>
      </div>

      <ViewConfigModal
        open={showModal}
        onOpenChange={setShowModal}
        agentName={config.agent_name || 'Unknown Agent'}
        level="Employee"
        config={config.config_override}
        isEnabled={config.is_enabled}
        updatedAt={config.updated_at}
      />
    </>
  );
}

type ResolvedConfigCardProps = {
  config: ResolvedAgentConfig;
  teamConfigs: TeamAgentConfig[];
  employeeConfigs: EmployeeAgentConfig[];
};

function ResolvedConfigCard({ config, teamConfigs, employeeConfigs }: ResolvedConfigCardProps) {
  const [showModal, setShowModal] = useState(false);

  // Determine source for each config key
  const getSource = (key: string): 'org' | 'team' | 'employee' => {
    const employeeConfig = employeeConfigs.find(c => c.agent_id === config.agent_id);
    if (employeeConfig && key in employeeConfig.config_override) {
      return 'employee';
    }
    const teamConfig = teamConfigs.find(c => c.agent_id === config.agent_id);
    if (teamConfig && key in teamConfig.config_override) {
      return 'team';
    }
    return 'org';
  };

  const configEntries = Object.entries(config.config).slice(0, 3);

  return (
    <>
      <div className="flex items-start justify-between p-4 border rounded-lg bg-card">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <h4 className="font-medium">{config.agent_name}</h4>
            <Badge variant={config.is_enabled ? 'default' : 'secondary'} className="text-xs">
              {config.is_enabled ? 'Enabled' : 'Disabled'}
            </Badge>
          </div>
          <div className="space-y-1">
            {configEntries.map(([key, value]) => {
              const source = getSource(key);
              const sourceLabel = source === 'org' ? '(org)' :
                source === 'team' ? '(team override)' : '(employee override)';
              const sourceColor = source === 'org' ? 'text-muted-foreground' :
                source === 'team' ? 'text-blue-500' : 'text-purple-500';

              return (
                <p key={key} className="text-sm font-mono">
                  {key}: {typeof value === 'object' ? JSON.stringify(value) : String(value)}{' '}
                  <span className={`text-xs ${sourceColor}`}>{sourceLabel}</span>
                </p>
              );
            })}
          </div>
        </div>
        <Button variant="ghost" size="sm" onClick={() => setShowModal(true)}>
          View Full Config
        </Button>
      </div>

      <ViewConfigModal
        open={showModal}
        onOpenChange={setShowModal}
        agentName={config.agent_name}
        level="Resolved"
        config={config.config}
        isEnabled={config.is_enabled}
        updatedAt={null}
      />
    </>
  );
}

export function ConfigurationHierarchy({
  orgConfigs,
  teamConfigs,
  employeeConfigs,
  resolvedConfigs,
  teamName,
  employeeName,
}: ConfigurationHierarchyProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Configuration Hierarchy</CardTitle>
        <CardDescription>
          View how agent configurations are inherited and merged (Org → Team → Employee)
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Organization Configs */}
        <ConfigSection
          title="Organization Configs (Base)"
          subtitle="Base configurations that apply organization-wide"
          count={orgConfigs.length}
          icon={<Building2 className="h-5 w-5 text-muted-foreground" />}
        >
          {orgConfigs.length > 0 ? (
            orgConfigs.map((config) => (
              <OrgConfigCard key={config.id} config={config} />
            ))
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">
              No organization-level agent configs configured
            </p>
          )}
        </ConfigSection>

        {/* Team Configs */}
        <ConfigSection
          title={`Team Configs${teamName ? ` (${teamName} Overrides)` : ''}`}
          subtitle={teamName ? 'Team-level overrides that apply to this team' : 'Employee not assigned to a team'}
          count={teamConfigs.length}
          icon={<Users className="h-5 w-5 text-muted-foreground" />}
          defaultOpen={teamConfigs.length > 0}
        >
          {!teamName ? (
            <p className="text-sm text-muted-foreground text-center py-4">
              Employee not assigned to a team. Only organization configs will be applied.
            </p>
          ) : teamConfigs.length > 0 ? (
            teamConfigs.map((config) => (
              <TeamConfigCard key={config.id} config={config} />
            ))
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">
              No team-level overrides for this team. Using organization defaults.
            </p>
          )}
        </ConfigSection>

        {/* Employee Configs */}
        <ConfigSection
          title={`Employee Configs (${employeeName}'s Overrides)`}
          subtitle="Personal overrides specific to this employee"
          count={employeeConfigs.length}
          icon={<User className="h-5 w-5 text-muted-foreground" />}
          defaultOpen={employeeConfigs.length > 0}
        >
          {employeeConfigs.length > 0 ? (
            employeeConfigs.map((config) => (
              <EmployeeConfigCard key={config.id} config={config} />
            ))
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">
              No personal overrides for this employee
            </p>
          )}
        </ConfigSection>

        {/* Resolved Configs */}
        <ConfigSection
          title="Resolved Configs (Final Merged)"
          subtitle="The complete merged configuration that this employee receives"
          count={resolvedConfigs.length}
          icon={<CheckCircle2 className="h-5 w-5 text-green-600" />}
        >
          {resolvedConfigs.length > 0 ? (
            resolvedConfigs.map((config) => (
              <ResolvedConfigCard
                key={config.agent_id}
                config={config}
                teamConfigs={teamConfigs}
                employeeConfigs={employeeConfigs}
              />
            ))
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">
              No agents are enabled for this employee
            </p>
          )}
        </ConfigSection>
      </CardContent>
    </Card>
  );
}
