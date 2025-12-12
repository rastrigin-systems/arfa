'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useToast } from '@/components/ui/use-toast';
import { useRoles } from '@/lib/hooks/useRoles';
import { useTeams } from '@/lib/hooks/useTeams';
import type { components } from '@/lib/api/schema';

type Employee = components['schemas']['Employee'];
type UpdateEmployeeRequest = components['schemas']['UpdateEmployeeRequest'];

// Validation schema for employee creation
const createEmployeeSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Invalid email format')
    .max(255, 'Email must be less than 255 characters'),
  full_name: z
    .string()
    .min(2, 'Name must be at least 2 characters')
    .max(255, 'Name must be less than 255 characters'),
  team_id: z.string().uuid('Invalid team ID').optional().nullable().or(z.literal(null)),
  role_id: z.string().uuid('Role is required').min(1, 'Role is required'),
});

// Validation schema for employee update (email is read-only)
const updateEmployeeSchema = z.object({
  full_name: z
    .string()
    .min(2, 'Name must be at least 2 characters')
    .max(255, 'Name must be less than 255 characters'),
  team_id: z.string().uuid('Invalid team ID').optional().nullable().or(z.literal(null)),
  role_id: z.string().uuid('Role is required').min(1, 'Role is required'),
  status: z.enum(['active', 'inactive', 'suspended']).optional(),
});

type CreateEmployeeFormData = z.infer<typeof createEmployeeSchema>;
type UpdateEmployeeFormData = z.infer<typeof updateEmployeeSchema>;

interface EmployeeFormProps {
  employee?: Employee;
  mode: 'create' | 'edit';
  onSuccess?: (employee: Employee) => void;
}

export function EmployeeForm({ employee, mode, onSuccess }: EmployeeFormProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [temporaryPassword, setTemporaryPassword] = useState<string | null>(null);

  // Use React Query hooks for data fetching (includes auth automatically)
  const { data: roles = [], isLoading: rolesLoading } = useRoles();
  const { data: teams = [], isLoading: teamsLoading } = useTeams();
  const isLoadingOptions = rolesLoading || teamsLoading;

  const isEditMode = mode === 'edit';

  // Initialize form with appropriate schema and default values
  const form = useForm<CreateEmployeeFormData | UpdateEmployeeFormData>({
    resolver: zodResolver(isEditMode ? updateEmployeeSchema : createEmployeeSchema),
    defaultValues: isEditMode && employee
      ? {
          full_name: employee.full_name,
          team_id: employee.team_id || null,
          role_id: employee.role_id,
          status: employee.status,
        }
      : {
          email: '',
          full_name: '',
          team_id: null,
          role_id: '',
        },
  });

  const onSubmit = async (data: CreateEmployeeFormData | UpdateEmployeeFormData) => {
    setIsSubmitting(true);
    setTemporaryPassword(null);

    try {
      if (isEditMode && employee) {
        // Update employee using Next.js API route (handles auth forwarding)
        const updateData: UpdateEmployeeRequest = {
          full_name: data.full_name,
          team_id: data.team_id || null,
          role_id: data.role_id,
          status: (data as UpdateEmployeeFormData).status,
        };

        const response = await fetch(`/api/employees/${employee.id}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify(updateData),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || 'Failed to update employee');
        }

        const result = await response.json();

        toast({
          title: 'Success',
          description: 'Employee updated successfully',
        });

        if (onSuccess) {
          onSuccess(result as Employee);
        } else {
          router.push(`/employees/${employee.id}`);
        }
      } else {
        // Create employee using Next.js API route (handles auth forwarding)
        const createData = data as CreateEmployeeFormData;

        const response = await fetch('/api/employees', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({
            email: createData.email,
            full_name: createData.full_name,
            team_id: createData.team_id || undefined,
            role_id: createData.role_id,
            preferences: {},
          }),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || 'Failed to create employee');
        }

        const result = await response.json();
        const newEmployee = result.employee;
        const tempPassword = result.temporary_password;

        // Store temporary password to display
        setTemporaryPassword(tempPassword);

        toast({
          title: 'Success',
          description: 'Employee created successfully',
        });

        if (onSuccess) {
          onSuccess(newEmployee);
        } else {
          // Show success with password, then redirect after a delay
          setTimeout(() => {
            router.push(`/employees/${newEmployee.id}`);
          }, 5000);
        }
      }
    } catch (error) {
      console.error('Form submission error:', error);

      toast({
        variant: 'destructive',
        title: 'Error',
        description:
          error instanceof Error
            ? error.message
            : `Failed to ${isEditMode ? 'update' : 'create'} employee`,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    if (isEditMode && employee) {
      router.push(`/employees/${employee.id}`);
    } else {
      router.push('/employees');
    }
  };

  const copyPassword = () => {
    if (temporaryPassword) {
      navigator.clipboard.writeText(temporaryPassword);
      toast({
        title: 'Copied',
        description: 'Temporary password copied to clipboard',
      });
    }
  };

  // Show temporary password modal after successful creation
  if (temporaryPassword) {
    return (
      <div className="space-y-6">
        <div className="rounded-lg border border-green-200 bg-green-50 p-6">
          <h2 className="text-xl font-semibold text-green-900 mb-4">
            Employee Created Successfully!
          </h2>
          <div className="space-y-4">
            <div>
              <p className="text-sm font-medium text-green-800 mb-2">
                Temporary Password (copy and share securely):
              </p>
              <div className="flex items-center gap-2">
                <code className="flex-1 rounded bg-white px-3 py-2 font-mono text-sm border border-green-300">
                  {temporaryPassword}
                </code>
                <Button onClick={copyPassword} variant="outline" size="sm">
                  Copy Password
                </Button>
              </div>
            </div>
            <p className="text-sm text-green-700">
              The employee should change this password on their first login.
            </p>
            <Button onClick={() => router.push('/employees')} className="mt-4">
              Go to Employee List
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
        {/* Email Field (only on create) */}
        {!isEditMode && (
          <FormField
            control={form.control}
            name="email"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Email</FormLabel>
                <FormControl>
                  <Input
                    type="email"
                    placeholder="employee@example.com"
                    {...field}
                    disabled={isSubmitting}
                  />
                </FormControl>
                <FormDescription>
                  The employee&apos;s email address (used for login)
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        )}

        {/* Email Field (read-only on edit) */}
        {isEditMode && employee && (
          <div className="space-y-2">
            <FormLabel>Email</FormLabel>
            <Input
              type="email"
              value={employee.email}
              disabled
              className="bg-muted"
            />
            <p className="text-sm text-muted-foreground">
              Email cannot be changed
            </p>
          </div>
        )}

        {/* Full Name Field */}
        <FormField
          control={form.control}
          name="full_name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Full Name</FormLabel>
              <FormControl>
                <Input
                  placeholder="John Doe"
                  {...field}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>
                The employee&apos;s full name
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Team Field */}
        <FormField
          control={form.control}
          name="team_id"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Team (Optional)</FormLabel>
              <Select
                onValueChange={(value) => field.onChange(value === '__none__' ? null : value)}
                value={field.value ?? '__none__'}
                disabled={isSubmitting || isLoadingOptions}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a team" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="__none__">No Team</SelectItem>
                  {teams.map((team) => (
                    <SelectItem key={team.id} value={team.id}>
                      {team.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormDescription>
                Assign the employee to a team
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Role Field */}
        <FormField
          control={form.control}
          name="role_id"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Role</FormLabel>
              <Select
                onValueChange={field.onChange}
                defaultValue={field.value}
                disabled={isSubmitting || isLoadingOptions}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a role" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {roles.map((role) => (
                    <SelectItem key={role.id} value={role.id}>
                      {role.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormDescription>
                The employee&apos;s role determines their permissions
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Status Field (only on edit) */}
        {isEditMode && (
          <FormField
            control={form.control}
            name="status"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Status</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value}
                  disabled={isSubmitting}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Select status" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="suspended">Suspended</SelectItem>
                    <SelectItem value="inactive">Inactive</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>
                  Change the employee&apos;s account status
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        )}

        {/* Form Actions */}
        <div className="flex items-center justify-end gap-4 pt-4">
          <Button
            type="button"
            variant="outline"
            onClick={handleCancel}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting || isLoadingOptions}>
            {isSubmitting
              ? isEditMode
                ? 'Updating...'
                : 'Creating...'
              : isEditMode
              ? 'Update Employee'
              : 'Create Employee'}
          </Button>
        </div>
      </form>
    </Form>
  );
}
