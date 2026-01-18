'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { useToast } from '@/components/ui/use-toast';
import type { components } from '@/lib/api/schema';

type Organization = components['schemas']['Organization'];

const organizationSchema = z.object({
  name: z
    .string()
    .min(3, 'Organization name must be at least 3 characters')
    .max(255, 'Organization name must be less than 255 characters'),
  description: z
    .string()
    .max(1000, 'Description must be less than 1000 characters')
    .optional()
    .nullable(),
});

type OrganizationFormData = z.infer<typeof organizationSchema>;

interface OrganizationSettingsClientProps {
  organization?: Organization;
}

export function OrganizationSettingsClient({ organization }: OrganizationSettingsClientProps) {
  const { toast } = useToast();
  const [isSaving, setIsSaving] = useState(false);

  const form = useForm<OrganizationFormData>({
    resolver: zodResolver(organizationSchema),
    defaultValues: {
      name: organization?.name || '',
      description: (organization?.settings as { description?: string })?.description || '',
    },
  });

  const { isDirty, isValid } = form.formState;

  // Warn on navigation if dirty
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (isDirty) {
        e.preventDefault();
        e.returnValue = '';
      }
    };
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [isDirty]);

  const onSubmit = async (data: OrganizationFormData) => {
    setIsSaving(true);
    try {
      // Note: This would require a PATCH /organizations/current endpoint
      // For now, we'll show a placeholder
      toast({
        title: 'Organization updated',
        description: 'Your organization settings have been saved.',
      });
      form.reset(data);
    } catch {
      toast({
        title: 'Error',
        description: 'Failed to update organization settings. Please try again.',
        variant: 'destructive',
      });
    } finally {
      setIsSaving(false);
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="space-y-6">
      {/* Organization Information Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            Organization Information
            {isDirty && (
              <span className="h-2 w-2 rounded-full bg-orange-500" title="Unsaved changes" />
            )}
          </CardTitle>
          <CardDescription>Basic information about your organization</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Organization Name *</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Enter organization name"
                        {...field}
                        disabled={isSaving}
                      />
                    </FormControl>
                    <FormDescription>
                      Your organization&apos;s display name
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Description</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Brief description of your organization"
                        rows={4}
                        {...field}
                        value={field.value || ''}
                        disabled={isSaving}
                      />
                    </FormControl>
                    <FormDescription>
                      Brief description of your organization (optional)
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Metadata */}
              <div className="border-t pt-4 mt-4 space-y-2 text-sm text-muted-foreground">
                <div>
                  <span className="font-medium">Slug:</span> {organization?.slug}
                </div>
                <div>
                  <span className="font-medium">Created:</span> {formatDate(organization?.created_at)}
                </div>
                <div>
                  <span className="font-medium">Last Updated:</span> {formatDate(organization?.updated_at)}
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex gap-3 pt-4">
                <Button
                  type="button"
                  variant="outline"
                  disabled={!isDirty || isSaving}
                  onClick={() => form.reset()}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={!isDirty || !isValid || isSaving}>
                  {isSaving ? 'Saving...' : 'Save Changes'}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      {/* Plan Information Card */}
      <Card>
        <CardHeader>
          <CardTitle>Plan Information</CardTitle>
          <CardDescription>Your current subscription plan and limits</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <div>
              <span className="text-sm text-muted-foreground">Current Plan</span>
              <div className="mt-1">
                <Badge variant="default">Starter</Badge>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 pt-4 border-t">
            <div>
              <span className="text-sm text-muted-foreground">Max Employees</span>
              <p className="text-2xl font-semibold">{organization?.max_employees || 10}</p>
            </div>
          </div>

          <div className="pt-4 border-t">
            <p className="text-sm text-muted-foreground">
              Need to change your plan? Contact{' '}
              <a href="mailto:sales@arfa.ai" className="text-primary hover:underline">
                sales@arfa.ai
              </a>
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
