import { apiClient } from '@/lib/api/client';
import { getServerToken } from '@/lib/auth';
import { notFound } from 'next/navigation';
import { EmployeeDetailClient } from './EmployeeDetailClient';

type PageProps = {
  params: {
    id: string;
  };
};

export default async function EmployeeDetailPage({ params }: PageProps) {
  const token = await getServerToken();

  if (!token) {
    throw new Error('Unauthorized');
  }

  // Fetch employee details
  const { data: employee, error } = await apiClient.GET('/employees/{employee_id}', {
    params: {
      path: {
        employee_id: params.id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error || !employee) {
    notFound();
  }

  // Fetch all config data in parallel
  const [
    { data: employeeConfigsData },
    { data: orgConfigsData },
    { data: resolvedConfigsData },
  ] = await Promise.all([
    // Employee's agent configurations (overrides)
    apiClient.GET('/employees/{employee_id}/agent-configs', {
      params: {
        path: {
          employee_id: params.id,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }),
    // Organization-level agent configs
    apiClient.GET('/organizations/current/agent-configs', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }),
    // Resolved configs (merged result)
    apiClient.GET('/employees/{employee_id}/agent-configs/resolved', {
      params: {
        path: {
          employee_id: params.id,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }),
  ]);

  // Fetch team configs only if employee has a team
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let teamConfigs: any[] = [];
  if (employee.team_id) {
    const { data: teamConfigsData } = await apiClient.GET('/teams/{team_id}/agent-configs', {
      params: {
        path: {
          team_id: employee.team_id,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    teamConfigs = teamConfigsData?.configs || [];
  }

  return (
    <EmployeeDetailClient
      employee={employee}
      agentConfigs={employeeConfigsData?.configs || []}
      orgConfigs={orgConfigsData?.configs || []}
      teamConfigs={teamConfigs}
      resolvedConfigs={resolvedConfigsData?.configs || []}
    />
  );
}
