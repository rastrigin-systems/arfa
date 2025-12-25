'use client';

import { useFormState, useFormStatus } from 'react-dom';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { forgotPasswordAction, type ForgotPasswordFormState } from './actions';

function SubmitButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" className="w-full" disabled={pending}>
      {pending ? 'Sending...' : 'Send reset link'}
    </Button>
  );
}

export default function ForgotPasswordPage() {
  const initialState: ForgotPasswordFormState = {};
  const [state, formAction] = useFormState(forgotPasswordAction, initialState);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-center text-3xl">Arfa Enterprise</CardTitle>
          <CardTitle className="text-center text-2xl font-semibold">Forgot Password</CardTitle>
          <CardDescription className="text-center">
            Enter your email address and we&apos;ll send you a link to reset your password.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form action={formAction} className="space-y-4">
            {/* Success message */}
            {state.success && state.message && (
              <div
                role="alert"
                aria-live="polite"
                aria-describedby="reset-success"
                className="rounded-md bg-green-50 p-4 text-sm text-green-800 dark:bg-green-900/20 dark:text-green-400"
              >
                <div className="flex items-start gap-2">
                  <span className="text-lg">✅</span>
                  <div id="reset-success" className="space-y-2">
                    <p className="font-semibold">Password reset link sent to your email</p>
                    <p className="text-sm">
                      If an account exists with this email, you will receive a password reset link
                      within a few minutes.
                    </p>
                    <p className="text-sm">Please check your spam folder if you don&apos;t see it.</p>
                  </div>
                </div>
              </div>
            )}

            {/* Only show form fields if not yet successful */}
            {!state.success && (
              <>
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
                    aria-label="Email address"
                    aria-invalid={!!state.errors?.email}
                    aria-describedby={state.errors?.email ? 'email-error' : undefined}
                  />
                  {state.errors?.email && (
                    <p id="email-error" role="alert" className="text-sm text-destructive">
                      {state.errors.email.join(', ')}
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
              </>
            )}

            {/* Back to login link */}
            <div className="text-center text-sm">
              <Link href="/login" className="text-primary hover:underline">
                Back to login
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
