'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useFormState, useFormStatus } from 'react-dom';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { PasswordStrengthIndicator } from '@/components/auth/PasswordStrengthIndicator';
import { resetPasswordAction, verifyResetToken, type ResetPasswordFormState } from './actions';

function SubmitButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" className="w-full" disabled={pending}>
      {pending ? 'Resetting...' : 'Reset Password'}
    </Button>
  );
}

export default function ResetPasswordPage({ params }: { params: { token: string } }) {
  const router = useRouter();
  const [tokenValid, setTokenValid] = useState<boolean | null>(null);
  const [tokenError, setTokenError] = useState<string>('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordMismatch, setPasswordMismatch] = useState(false);

  const initialState: ResetPasswordFormState = {};
  const [state, formAction] = useFormState(resetPasswordAction, initialState);

  // Verify token on page load
  useEffect(() => {
    async function checkToken() {
      const result = await verifyResetToken(params.token);
      setTokenValid(result.valid);
      if (!result.valid) {
        setTokenError(result.error || 'Invalid token');
      }
    }

    checkToken();
  }, [params.token]);

  // Redirect to login after successful password reset
  useEffect(() => {
    if (state.success) {
      const timer = setTimeout(() => {
        router.push('/login');
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [state.success, router]);

  // Check password confirmation match in real-time
  useEffect(() => {
    if (confirmPassword && newPassword) {
      setPasswordMismatch(newPassword !== confirmPassword);
    } else {
      setPasswordMismatch(false);
    }
  }, [newPassword, confirmPassword]);

  // Loading state
  if (tokenValid === null) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <Card className="w-full max-w-lg">
          <CardContent className="pt-6">
            <p className="text-center text-muted-foreground">Verifying reset link...</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Invalid/expired/used token error state
  if (!tokenValid) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <Card className="w-full max-w-lg">
          <CardHeader>
            <CardTitle className="text-center text-3xl">Arfa</CardTitle>
            <CardTitle className="text-center text-2xl font-semibold">Reset Password</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              role="alert"
              className="rounded-md bg-destructive/10 p-4 text-sm text-destructive"
            >
              <div className="flex items-start gap-2">
                <span className="text-lg">⚠️</span>
                <div className="space-y-2">
                  <p className="font-semibold">This reset link has expired or is invalid</p>
                  <p className="text-sm">
                    {tokenError.includes('expired')
                      ? 'Password reset links are valid for 1 hour.'
                      : tokenError.includes('used')
                        ? 'Password reset links can only be used once.'
                        : 'This link is invalid or has already been used.'}
                  </p>
                </div>
              </div>
            </div>

            <div className="mt-4 text-center">
              <Link href="/forgot-password">
                <Button variant="outline" className="w-full">
                  Request new reset link
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Success state
  if (state.success) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <Card className="w-full max-w-lg">
          <CardHeader>
            <CardTitle className="text-center text-3xl">Arfa</CardTitle>
            <CardTitle className="text-center text-2xl font-semibold">Reset Password</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              role="alert"
              aria-live="polite"
              className="rounded-md bg-green-50 p-4 text-sm text-green-800 dark:bg-green-900/20 dark:text-green-400"
            >
              <div className="flex items-start gap-2">
                <span className="text-lg">✅</span>
                <div className="space-y-2">
                  <p className="font-semibold">Password reset successful!</p>
                  <p className="text-sm">Redirecting to login...</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Reset password form
  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle className="text-center text-3xl">Arfa</CardTitle>
          <CardTitle className="text-center text-2xl font-semibold">Reset Password</CardTitle>
        </CardHeader>
        <CardContent>
          <form action={formAction} className="space-y-4">
            {/* Hidden token field */}
            <input type="hidden" name="token" value={params.token} />

            {/* New Password field */}
            <div className="space-y-2">
              <Label htmlFor="new_password">New Password</Label>
              <Input
                id="new_password"
                name="new_password"
                type="password"
                placeholder="••••••••"
                required
                aria-required="true"
                aria-label="New password"
                aria-invalid={!!state.errors?.new_password}
                aria-describedby={
                  state.errors?.new_password ? 'new-password-error password-requirements' : 'password-requirements'
                }
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
              />
              <PasswordStrengthIndicator password={newPassword} />
              {state.errors?.new_password && (
                <p id="new-password-error" role="alert" className="text-sm text-destructive">
                  {state.errors.new_password.join(', ')}
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
                aria-label="Confirm new password"
                aria-invalid={!!state.errors?.confirm_password || passwordMismatch}
                aria-describedby={state.errors?.confirm_password ? 'confirm-password-error' : undefined}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className={passwordMismatch ? 'border-destructive' : ''}
              />
              {passwordMismatch && (
                <p role="alert" aria-live="polite" className="text-sm text-destructive">
                  ✗ Passwords do not match
                </p>
              )}
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
            <SubmitButton />
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
