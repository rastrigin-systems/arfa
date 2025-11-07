'use client';

import { useState } from 'react';
import { Users, UserPlus, Mail, Clock, CheckCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { useEmployees } from '@/lib/hooks/useEmployees';
import { useInvitations } from '@/lib/hooks/useInvitations';
import { InviteEmployeeModal } from '@/components/team/InviteEmployeeModal';

export default function TeamManagementPage() {
  const [activeTab, setActiveTab] = useState('employees');
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [inviteModalOpen, setInviteModalOpen] = useState(false);

  // Fetch employees
  const {
    data: employeesData,
    isLoading: employeesLoading,
    error: employeesError,
  } = useEmployees({ page, limit: 10, search });

  // Fetch invitations
  const {
    data: invitationsData,
    isLoading: invitationsLoading,
    error: invitationsError,
  } = useInvitations({ page, limit: 10, status: 'pending' });

  const totalEmployees = employeesData?.total || 0;
  const activeEmployees =
    employeesData?.employees.filter((e) => e.status === 'active').length || 0;
  const pendingInvites = invitationsData?.total || 0;

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <Users className="h-8 w-8" />
          Team Management
        </h1>
        <p className="text-muted-foreground mt-2">
          Manage your organization&apos;s employees and invitations
        </p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Employees</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalEmployees}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activeEmployees}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Pending Invites</CardTitle>
            <Mail className="h-4 w-4 text-yellow-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{pendingInvites}</div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Card>
        <CardContent className="pt-6">
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <div className="flex items-center justify-between mb-4">
              <TabsList>
                <TabsTrigger value="employees">Employees</TabsTrigger>
                <TabsTrigger value="invitations">
                  Invitations
                  {pendingInvites > 0 && (
                    <Badge variant="secondary" className="ml-2">
                      {pendingInvites}
                    </Badge>
                  )}
                </TabsTrigger>
              </TabsList>

              <div className="flex items-center gap-2">
                <Input
                  type="search"
                  placeholder="Search by name or email..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="w-64"
                  aria-label="Search employees"
                />
                <Button onClick={() => setInviteModalOpen(true)}>
                  <UserPlus className="h-4 w-4 mr-2" />
                  Invite Employee
                </Button>
              </div>
            </div>

            {/* Employees Tab */}
            <TabsContent value="employees" className="mt-0">
              {employeesLoading ? (
                <div className="flex justify-center py-8" role="status" aria-label="Loading employees">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
                </div>
              ) : employeesError ? (
                <div className="text-center py-8 text-destructive" role="alert">
                  <p>Failed to load employees. Please try again.</p>
                </div>
              ) : employeesData?.employees.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <Users className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>No employees found</p>
                </div>
              ) : (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Email</TableHead>
                        <TableHead>Role</TableHead>
                        <TableHead>Team</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {employeesData?.employees.map((employee) => (
                        <TableRow key={employee.id}>
                          <TableCell className="font-medium">
                            <div>
                              <div>{employee.full_name}</div>
                              <div className="text-xs text-muted-foreground">
                                Joined{' '}
                                {new Date(employee.created_at).toLocaleDateString()}
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>{employee.email}</TableCell>
                          <TableCell>
                            <Badge variant="outline">{employee.role.name}</Badge>
                          </TableCell>
                          <TableCell>
                            {employee.team ? employee.team.name : 'No team'}
                          </TableCell>
                          <TableCell>
                            <Badge
                              variant={
                                employee.status === 'active'
                                  ? 'default'
                                  : employee.status === 'inactive'
                                  ? 'secondary'
                                  : 'destructive'
                              }
                            >
                              {employee.status}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="sm">
                              •••
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}

              {/* Pagination */}
              {employeesData && employeesData.total > 10 && (
                <div className="flex items-center justify-between mt-4">
                  <p className="text-sm text-muted-foreground">
                    Showing {(page - 1) * 10 + 1} to{' '}
                    {Math.min(page * 10, employeesData.total)} of {employeesData.total}{' '}
                    employees
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      disabled={page === 1}
                    >
                      Previous
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => p + 1)}
                      disabled={page * 10 >= employeesData.total}
                    >
                      Next
                    </Button>
                  </div>
                </div>
              )}
            </TabsContent>

            {/* Invitations Tab */}
            <TabsContent value="invitations" className="mt-0">
              {invitationsLoading ? (
                <div className="flex justify-center py-8" role="status" aria-label="Loading invitations">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
                </div>
              ) : invitationsError ? (
                <div className="text-center py-8 text-destructive" role="alert">
                  <p>Failed to load invitations. Please try again.</p>
                </div>
              ) : invitationsData?.invitations.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <Mail className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>No pending invitations</p>
                </div>
              ) : (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Email</TableHead>
                        <TableHead>Invited By</TableHead>
                        <TableHead>Role</TableHead>
                        <TableHead>Team</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Details</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {invitationsData?.invitations.map((invitation) => {
                        const expiresAt = new Date(invitation.expires_at);
                        const now = new Date();
                        const daysUntilExpiry = Math.ceil(
                          (expiresAt.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
                        );

                        return (
                          <TableRow key={invitation.id}>
                            <TableCell className="font-medium">
                              <div className="flex items-center gap-2">
                                <Mail className="h-4 w-4 text-muted-foreground" />
                                {invitation.email}
                              </div>
                            </TableCell>
                            <TableCell>{invitation.inviter.full_name}</TableCell>
                            <TableCell>
                              <Badge variant="outline">{invitation.role.name}</Badge>
                            </TableCell>
                            <TableCell>
                              {invitation.team ? invitation.team.name : 'No team'}
                            </TableCell>
                            <TableCell>
                              <Badge
                                variant={
                                  invitation.status === 'pending'
                                    ? 'secondary'
                                    : invitation.status === 'accepted'
                                    ? 'default'
                                    : 'destructive'
                                }
                              >
                                {invitation.status === 'pending' && daysUntilExpiry <= 1 && (
                                  <Clock className="h-3 w-3 mr-1" />
                                )}
                                {invitation.status}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-sm text-muted-foreground">
                              <div>
                                Sent {new Date(invitation.created_at).toLocaleDateString()}
                              </div>
                              <div>
                                {daysUntilExpiry > 0
                                  ? `Expires in ${daysUntilExpiry} day${
                                      daysUntilExpiry > 1 ? 's' : ''
                                    }`
                                  : 'Expires today'}
                              </div>
                            </TableCell>
                            <TableCell className="text-right">
                              <Button variant="ghost" size="sm">
                                •••
                              </Button>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </div>
              )}

              {/* Pagination */}
              {invitationsData && invitationsData.total > 10 && (
                <div className="flex items-center justify-between mt-4">
                  <p className="text-sm text-muted-foreground">
                    Showing {(page - 1) * 10 + 1} to{' '}
                    {Math.min(page * 10, invitationsData.total)} of{' '}
                    {invitationsData.total} invitations
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      disabled={page === 1}
                    >
                      Previous
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => p + 1)}
                      disabled={page * 10 >= invitationsData.total}
                    >
                      Next
                    </Button>
                  </div>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Invite Modal */}
      <InviteEmployeeModal
        open={inviteModalOpen}
        onOpenChange={setInviteModalOpen}
      />
    </div>
  );
}
