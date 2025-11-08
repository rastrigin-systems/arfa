'use client';

import { useFormState, useFormStatus } from 'react-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { loginAction, type LoginFormState } from './actions';

function SubmitButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" className="w-full" disabled={pending}>
      {pending ? 'Logging in...' : 'Login'}
    </Button>
  );
}

export default function LoginPage() {
  const initialState: LoginFormState = {};
  const [state, formAction] = useFormState(loginAction, initialState);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-center text-3xl">Ubik Enterprise</CardTitle>
          <CardDescription className="text-center">
            Sign in to manage your AI agent configurations
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form action={formAction} className="space-y-4">
            {/* Email field */}
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                name="email"
                type="email"
                placeholder="you@example.com"
                defaultValue="alice@acme.com"
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

            {/* Password field */}
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                name="password"
                type="password"
                placeholder="••••••••"
                defaultValue="SecurePass123!"
                required
                aria-required="true"
                aria-invalid={!!state.errors?.password}
                aria-describedby={state.errors?.password ? 'password-error' : undefined}
              />
              {state.errors?.password && (
                <p id="password-error" role="alert" className="text-sm text-destructive">
                  {state.errors.password.join(', ')}
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
            <SubmitButton />

            {/* Sign up link */}
            <div className="text-center text-sm">
              Don&apos;t have an account?{' '}
              <a href="/signup" className="text-primary hover:underline">
                Sign up
              </a>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
