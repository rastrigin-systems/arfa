'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Info } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/components/ui/use-toast';
import { useRoles } from '@/lib/hooks/useRoles';
import { useTeams } from '@/lib/hooks/useTeams';
import { useCreateInvitation } from '@/lib/hooks/useInvitations';

const inviteSchema = z.object({
  email: z.string().email('Invalid email address'),
  role_id: z.string().min(1, 'Role is required'),
  team_id: z.string().optional(),
  message: z.string().max(500, 'Message must be 500 characters or less').optional(),
});

type InviteFormData = z.infer<typeof inviteSchema>;

interface InviteEmployeeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function InviteEmployeeModal({ open, onOpenChange }: InviteEmployeeModalProps) {
  const { toast } = useToast();
  const { data: roles, isLoading: rolesLoading } = useRoles();
  const { data: teams, isLoading: teamsLoading } = useTeams();
  const createInvitation = useCreateInvitation();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
    setValue,
    watch,
  } = useForm<InviteFormData>({
    resolver: zodResolver(inviteSchema),
  });

  const message = watch('message') || '';

  const onSubmit = async (data: InviteFormData) => {
    try {
      await createInvitation.mutateAsync({
        email: data.email,
        role_id: data.role_id,
        team_id: data.team_id || undefined,
        message: data.message || undefined,
      });

      toast({
        title: 'Invitation sent',
        description: `Invitation sent to ${data.email}`,
      });

      reset();
      onOpenChange(false);
    } catch (error) {
      toast({
        title: 'Failed to send invitation',
        description: error instanceof Error ? error.message : 'An error occurred',
        variant: 'destructive',
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Invite Employee to Join</DialogTitle>
          <DialogDescription>
            Invite a new employee to your organization via email.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {/* Email */}
          <div className="space-y-2">
            <Label htmlFor="email">
              Email Address <span className="text-destructive">*</span>
            </Label>
            <Input
              id="email"
              type="email"
              placeholder="colleague@company.com"
              {...register('email')}
              aria-invalid={!!errors.email}
              aria-describedby={errors.email ? 'email-error' : undefined}
            />
            {errors.email && (
              <p id="email-error" role="alert" className="text-sm text-destructive">
                {errors.email.message}
              </p>
            )}
          </div>

          {/* Role */}
          <div className="space-y-2">
            <Label htmlFor="role">
              Role <span className="text-destructive">*</span>
            </Label>
            <Select
              onValueChange={(value) => setValue('role_id', value)}
              disabled={rolesLoading}
            >
              <SelectTrigger id="role" aria-invalid={!!errors.role_id}>
                <SelectValue placeholder="Select a role" />
              </SelectTrigger>
              <SelectContent>
                {roles?.map((role) => (
                  <SelectItem key={role.id} value={role.id}>
                    {role.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.role_id && (
              <p role="alert" className="text-sm text-destructive">
                {errors.role_id.message}
              </p>
            )}
          </div>

          {/* Team (Optional) */}
          <div className="space-y-2">
            <Label htmlFor="team">Team (Optional)</Label>
            <Select
              onValueChange={(value) => setValue('team_id', value === '__none__' ? undefined : value)}
              disabled={teamsLoading}
            >
              <SelectTrigger id="team">
                <SelectValue placeholder="Select a team" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__none__">No team</SelectItem>
                {teams?.map((team) => (
                  <SelectItem key={team.id} value={team.id}>
                    {team.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Personal Message (Optional) */}
          <div className="space-y-2">
            <Label htmlFor="message">Personal Message (Optional)</Label>
            <Textarea
              id="message"
              placeholder="Add a personal note to your invitation (optional)"
              rows={3}
              maxLength={500}
              {...register('message')}
              aria-describedby="message-counter"
            />
            <p id="message-counter" className="text-xs text-muted-foreground text-right">
              {message.length} / 500 characters
            </p>
          </div>

          {/* Info Message */}
          <div className="flex gap-2 p-3 bg-muted rounded-md">
            <Info className="h-4 w-4 text-muted-foreground flex-shrink-0 mt-0.5" />
            <p className="text-sm text-muted-foreground">
              They&apos;ll receive an email with a secure link to accept the invitation and set up
              their account. The invitation expires in 7 days.
            </p>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                reset();
                onOpenChange(false);
              }}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Sending...' : 'Send Invite'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
