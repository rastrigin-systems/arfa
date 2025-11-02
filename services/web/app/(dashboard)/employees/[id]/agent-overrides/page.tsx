import { apiClient } from '@/lib/api/client';
import { getServerToken } from '@/lib/auth';
import { notFound } from 'next/navigation';
import { EmployeeAgentOverridesClient } from './EmployeeAgentOverridesClient';

type PageProps = {
  params: {
    id: string;
  };
};

export default async function EmployeeAgentOverridesPage({ params }: PageProps) {
  const token = await getServerToken();

  if (!token) {
    throw new Error('Unauthorized');
  }

  // Fetch employee details
  const { data: employee, error: employeeError } = await apiClient.GET('/employees/{employee_id}', {
    params: {
      path: {
        employee_id: params.id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (employeeError || !employee) {
    notFound();
  }

  // Fetch employee's agent configurations (overrides)
  const { data: agentConfigsData } = await apiClient.GET('/employees/{employee_id}/agent-configs', {
    params: {
      path: {
        employee_id: params.id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return (
    <EmployeeAgentOverridesClient
      employee={employee}
      agentOverrides={agentConfigsData?.configs || []}
    />
  );
}
