// Shared types for API responses

export interface Employee {
  id: string;
  email: string;
  full_name: string;
  status: 'active' | 'inactive' | 'suspended';
  role: {
    id: string;
    name: string;
  };
  team: {
    id: string;
    name: string;
  } | null;
  created_at: string;
}

export interface Role {
  id: string;
  name: string;
  description?: string;
}

export interface Team {
  id: string;
  name: string;
  description?: string;
}

export interface Invitation {
  id: string;
  email: string;
  role: {
    id: string;
    name: string;
  };
  team: {
    id: string;
    name: string;
  } | null;
  inviter: {
    id: string;
    full_name: string;
    email: string;
  };
  status: 'pending' | 'accepted' | 'expired' | 'cancelled';
  expires_at: string;
  created_at: string;
}

export interface EmployeesParams {
  page: number;
  limit: number;
  search?: string;
  team?: string;
  role?: string;
  status?: string;
}

export interface EmployeesResponse {
  employees: Employee[];
  total: number;
  page: number;
  limit: number;
}

export interface InvitationsParams {
  page: number;
  limit: number;
  status?: 'pending' | 'accepted' | 'expired' | 'cancelled';
}

export interface InvitationsResponse {
  invitations: Invitation[];
  total: number;
  page: number;
  limit: number;
}

export interface CreateInvitationParams {
  email: string;
  role_id: string;
  team_id?: string;
  message?: string;
}

export interface CreateInvitationResponse {
  invitation: {
    id: string;
    email: string;
    token: string;
    invitation_url: string;
    expires_at: string;
  };
}

export interface UpdateEmployeeParams {
  team_id?: string;
  role_id?: string;
  status?: 'active' | 'inactive' | 'suspended';
}
