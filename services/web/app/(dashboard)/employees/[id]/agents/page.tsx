import { notFound } from 'next/navigation';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { EmployeeAgentConfigsClient } from './EmployeeAgentConfigsClient';

type PageProps = {
  params: { id: string };
};

export default async function EmployeeAgentConfigsPage({ params }: PageProps) {
  const token = await getServerToken();

  if (!token) {
    return notFound();
  }

  // Fetch employee details
  const employeeResponse = await apiClient.GET('/employees/{employee_id}', {
    params: {
      path: { employee_id: params.id },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (employeeResponse.error || !employeeResponse.data) {
    return notFound();
  }

  // Fetch employee agent configs
  const agentConfigsResponse = await apiClient.GET('/employees/{employee_id}/agent-configs', {
    params: {
      path: { employee_id: params.id },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  // Fetch available agents from catalog
  const agentsResponse = await apiClient.GET('/agents', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  const employee = employeeResponse.data;
  const agentConfigs = agentConfigsResponse.data?.configs || [];
  const agents = agentsResponse.data?.agents || [];

  return (
    <EmployeeAgentConfigsClient
      employee={employee}
      initialConfigs={agentConfigs}
      initialAgents={agents}
    />
  );
}
