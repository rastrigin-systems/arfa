'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useFormState, useFormStatus } from 'react-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { PasswordStrengthIndicator } from '@/components/auth/PasswordStrengthIndicator';
import { SlugAvailabilityCheck } from '@/components/auth/SlugAvailabilityCheck';
import { signupAction, type SignupFormState } from './actions';

function SubmitButton({ disabled }: { disabled: boolean }) {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" className="w-full" disabled={pending || disabled}>
      {pending ? 'Creating Account...' : 'Create Account'}
    </Button>
  );
}

export default function SignupPage() {
  const initialState: SignupFormState = {};
  const [state, formAction] = useFormState(signupAction, initialState);

  const [password, setPassword] = useState('');
  const [orgSlug, setOrgSlug] = useState('');
  const [isSlugAvailable, setIsSlugAvailable] = useState(false);

  // Disable submit if slug is not available
  const isFormDisabled = orgSlug.length >= 3 && !isSlugAvailable;

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle className="text-center text-3xl">Ubik Enterprise</CardTitle>
          <CardDescription className="text-center">
            Create Your Organization Account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form action={formAction} className="space-y-4">
            {/* Full Name field */}
            <div className="space-y-2">
              <Label htmlFor="full_name">Full Name</Label>
              <Input
                id="full_name"
                name="full_name"
                type="text"
                placeholder="John Doe"
                required
                aria-required="true"
                aria-invalid={!!state.errors?.full_name}
                aria-describedby={state.errors?.full_name ? 'full-name-error' : undefined}
              />
              {state.errors?.full_name && (
                <p id="full-name-error" role="alert" className="text-sm text-destructive">
                  {state.errors.full_name.join(', ')}
                </p>
              )}
            </div>

            {/* Email field */}
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                name="email"
                type="email"
                placeholder="you@example.com"
                required
                aria-required="true"
                aria-invalid={!!state.errors?.email}
                aria-describedby={state.errors?.email ? 'email-error' : undefined}
              />
              {state.errors?.email && (
                <p id="email-error" role="alert" className="text-sm text-destructive">
                  {state.errors.email.join(', ')}
                </p>
              )}
            </div>

            {/* Organization Name field */}
            <div className="space-y-2">
              <Label htmlFor="org_name">Organization Name</Label>
              <Input
                id="org_name"
                name="org_name"
                type="text"
                placeholder="Acme Corporation"
                required
                aria-required="true"
                aria-invalid={!!state.errors?.org_name}
                aria-describedby={state.errors?.org_name ? 'org-name-error' : undefined}
              />
              {state.errors?.org_name && (
                <p id="org-name-error" role="alert" className="text-sm text-destructive">
                  {state.errors.org_name.join(', ')}
                </p>
              )}
            </div>

            {/* Organization Slug field */}
            <div className="space-y-2">
              <Label htmlFor="org_slug">Organization Slug</Label>
              <Input
                id="org_slug"
                name="org_slug"
                type="text"
                placeholder="acme"
                required
                aria-required="true"
                aria-invalid={!!state.errors?.org_slug}
                aria-describedby={state.errors?.org_slug ? 'org-slug-error' : undefined}
                value={orgSlug}
                onChange={(e) => setOrgSlug(e.target.value.toLowerCase())}
              />
              <SlugAvailabilityCheck
                slug={orgSlug}
                onAvailabilityChange={setIsSlugAvailable}
              />
              {state.errors?.org_slug && (
                <p id="org-slug-error" role="alert" className="text-sm text-destructive">
                  {state.errors.org_slug.join(', ')}
                </p>
              )}
            </div>

            {/* Password field */}
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                name="password"
                type="password"
                placeholder="••••••••"
                required
                aria-required="true"
                aria-invalid={!!state.errors?.password}
                aria-describedby={state.errors?.password ? 'password-error' : undefined}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
              <PasswordStrengthIndicator password={password} />
              {state.errors?.password && (
                <p id="password-error" role="alert" className="text-sm text-destructive">
                  {state.errors.password.join(', ')}
                </p>
              )}
            </div>

            {/* Confirm Password field */}
            <div className="space-y-2">
              <Label htmlFor="confirm_password">Confirm Password</Label>
              <Input
                id="confirm_password"
                name="confirm_password"
                type="password"
                placeholder="••••••••"
                required
                aria-required="true"
                aria-invalid={!!state.errors?.confirm_password}
                aria-describedby={state.errors?.confirm_password ? 'confirm-password-error' : undefined}
              />
              {state.errors?.confirm_password && (
                <p id="confirm-password-error" role="alert" className="text-sm text-destructive">
                  {state.errors.confirm_password.join(', ')}
                </p>
              )}
            </div>

            {/* Form-level errors */}
            {state.errors?._form && (
              <div role="alert" className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {state.errors._form.join(', ')}
              </div>
            )}

            {/* Submit button */}
            <SubmitButton disabled={isFormDisabled} />

            {/* Link to login */}
            <div className="text-center text-sm text-muted-foreground">
              Already have an account?{' '}
              <Link href="/login" className="text-primary hover:underline">
                Sign in
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
