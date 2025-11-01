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

  // Fetch employee's agent configurations
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
    <EmployeeDetailClient
      employee={employee}
      agentConfigs={agentConfigsData?.configs || []}
    />
  );
}
