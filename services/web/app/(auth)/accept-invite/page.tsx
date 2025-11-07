'use client';

import { Suspense, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { PasswordStrengthIndicator } from '@/components/auth/PasswordStrengthIndicator';
import { useInvitation, useAcceptInvitation } from '@/lib/hooks/useInvitation';
import { acceptInvitationSchema, type AcceptInvitationFormData } from '@/lib/validation/accept-invitation';
import { Building2, User, Target, Users, Clock, AlertCircle, CheckCircle2, Loader2 } from 'lucide-react';

function AcceptInviteContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const token = searchParams?.get('token');

  const { state, invitation, error: validationError } = useInvitation(token);
  const { acceptInvitation, isSubmitting, error: submissionError, errorCode } = useAcceptInvitation(token || '');

  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
    setValue,
  } = useForm<AcceptInvitationFormData>({
    resolver: zodResolver(acceptInvitationSchema),
    defaultValues: {
      full_name: '',
      password: '',
      confirm_password: '',
      terms_accepted: false,
    },
  });

  const termsAccepted = watch('terms_accepted');

  const onSubmit = async (data: AcceptInvitationFormData) => {
    try {
      const response = await acceptInvitation({
        full_name: data.full_name,
        password: data.password,
      });

      // Store session token
      localStorage.setItem('ubik_session_token', response.token);

      // Redirect to dashboard
      router.push('/dashboard');
    } catch (err) {
      // Error is handled by useAcceptInvitation hook
      console.error('Failed to accept invitation:', err);
    }
  };

  // Loading state
  if (state === 'loading') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <Card className="w-full max-w-lg">
          <CardContent className="py-8">
            <div
              role="status"
              aria-label="Validating invitation token"
              className="flex flex-col items-center justify-center space-y-4"
            >
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <p className="text-sm text-muted-foreground">Validating invitation...</p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Invalid token state
  if (state === 'invalid') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <Card className="w-full max-w-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-center text-2xl">
              <AlertCircle className="h-6 w-6 text-destructive" />
              Invalid Invitation
            </CardTitle>
            <CardDescription className="text-center">
              This invitation link is invalid or has expired.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-lg border p-4 space-y-2">
              <p className="text-sm font-medium">Possible reasons:</p>
              <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
                <li>The invitation has expired (7 days)</li>
                <li>The invitation was cancelled by the sender</li>
                <li>The invitation has already been accepted</li>
                <li>The link is malformed or incorrect</li>
              </ul>
            </div>

            <div className="space-y-2">
              <p className="text-sm font-medium">What to do next:</p>
              <ol className="list-decimal list-inside text-sm text-muted-foreground space-y-1">
                <li>Contact the person who invited you and request a new invitation</li>
                <li>If you already have an account, please log in</li>
              </ol>
            </div>

            <div className="flex gap-2">
              <Button asChild variant="outline" className="flex-1">
                <Link href="/login">Go to Login</Link>
              </Button>
              <Button asChild variant="outline" className="flex-1">
                <Link href="/support">Contact Support</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Expired token state
  if (state === 'expired') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <Card className="w-full max-w-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-center text-2xl">
              <Clock className="h-6 w-6 text-yellow-600" />
              Invitation Expired
            </CardTitle>
            <CardDescription className="text-center">
              This invitation has expired.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{validationError}</AlertDescription>
            </Alert>

            <div className="space-y-2">
              <p className="text-sm">Contact the person who invited you to request a new invitation link.</p>
            </div>

            <div className="flex gap-2">
              <Button asChild variant="outline" className="flex-1">
                <Link href="/login">Go to Login</Link>
              </Button>
              <Button asChild variant="outline" className="flex-1">
                <Link href="/support">Contact Support</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Already accepted state
  if (state === 'accepted') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <Card className="w-full max-w-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-center text-2xl">
              <CheckCircle2 className="h-6 w-6 text-green-600" />
              Invitation Already Accepted
            </CardTitle>
            <CardDescription className="text-center">
              This invitation has already been used.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-lg border p-4 space-y-2">
              <p className="text-sm">If you accepted this invitation, please log in to access your account.</p>
              <p className="text-sm text-muted-foreground">
                If you did not accept this invitation, please contact the organization administrator.
              </p>
            </div>

            <Button asChild className="w-full">
              <Link href="/login">Go to Login</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Error state
  if (state === 'error') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <Card className="w-full max-w-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-center text-2xl">
              <AlertCircle className="h-6 w-6 text-destructive" />
              Error Loading Invitation
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{validationError || 'An unexpected error occurred'}</AlertDescription>
            </Alert>

            <div className="flex gap-2">
              <Button asChild variant="outline" className="flex-1">
                <Link href="/login">Go to Login</Link>
              </Button>
              <Button asChild variant="outline" className="flex-1">
                <Link href="/support">Contact Support</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Valid state - show form
  if (!invitation) return null;

  const daysUntilExpiry = Math.ceil(
    (new Date(invitation.expires_at).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24)
  );
  const expiryColor = daysUntilExpiry > 3 ? 'text-green-600' : daysUntilExpiry > 1 ? 'text-yellow-600' : 'text-red-600';
  const expiryText = daysUntilExpiry === 1 ? 'Tomorrow' : `In ${daysUntilExpiry} days`;
  const absoluteDate = new Date(invitation.expires_at).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle className="text-center text-3xl">Ubik Enterprise</CardTitle>
          <CardDescription className="text-center text-lg">You&apos;ve Been Invited!</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Invitation Details */}
          <div className="rounded-lg border p-4 space-y-4">
            <h3 className="font-semibold text-lg">Invitation Details</h3>

            {/* Organization */}
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Organization</Label>
              <div className="flex items-center gap-2 p-3 rounded-md bg-muted">
                <Building2 className="h-5 w-5 text-primary" />
                <div>
                  <p className="font-medium">{invitation.organization.name}</p>
                  <p className="text-sm text-muted-foreground">https://{invitation.organization.slug}.ubik.io</p>
                </div>
              </div>
            </div>

            {/* Inviter */}
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Invited By</Label>
              <div className="flex items-center gap-2 p-3 rounded-md bg-muted">
                <User className="h-5 w-5 text-primary" />
                <div>
                  <p className="font-medium">
                    {invitation.inviter.full_name} ({invitation.inviter.email})
                  </p>
                  <p className="text-sm text-muted-foreground">{invitation.inviter.role.name}</p>
                </div>
              </div>
            </div>

            {/* Role */}
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Your Role</Label>
              <div className="flex items-center gap-2 p-3 rounded-md bg-muted">
                <Target className="h-5 w-5 text-primary" />
                <div>
                  <p className="font-medium">{invitation.role.name}</p>
                  <p className="text-sm text-muted-foreground">{invitation.role.description}</p>
                </div>
              </div>
            </div>

            {/* Team */}
            {invitation.team && (
              <div className="space-y-1">
                <Label className="text-xs text-muted-foreground">Team Assignment</Label>
                <div className="flex items-center gap-2 p-3 rounded-md bg-muted">
                  <Users className="h-5 w-5 text-primary" />
                  <p className="font-medium">{invitation.team.name}</p>
                </div>
              </div>
            )}

            {/* Expiration */}
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Invitation Expires</Label>
              <div className="flex items-center gap-2 p-3 rounded-md bg-muted">
                <Clock className="h-5 w-5 text-primary" />
                <p className={`font-medium ${expiryColor}`}>
                  {expiryText} ({absoluteDate})
                </p>
              </div>
            </div>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="rounded-lg border p-4 space-y-4">
              <h3 className="font-semibold text-lg">Set Your Password</h3>

              {/* Email (read-only) */}
              <div className="space-y-2">
                <Label htmlFor="email">Your Email</Label>
                <div className="relative">
                  <Input id="email" type="email" value={invitation.email} disabled className="pr-24" />
                  <Badge variant="secondary" className="absolute right-2 top-1/2 -translate-y-1/2">
                    <CheckCircle2 className="h-3 w-3 mr-1" />
                    Verified
                  </Badge>
                </div>
                <p className="text-xs text-muted-foreground">This email is associated with your invitation</p>
              </div>

              {/* Full Name */}
              <div className="space-y-2">
                <Label htmlFor="full_name">
                  Full Name <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="full_name"
                  type="text"
                  placeholder="Jane Doe"
                  {...register('full_name')}
                  aria-required="true"
                  aria-invalid={!!errors.full_name}
                  aria-describedby={errors.full_name ? 'full-name-error' : undefined}
                />
                {errors.full_name && (
                  <p id="full-name-error" role="alert" className="text-sm text-destructive">
                    {errors.full_name.message}
                  </p>
                )}
              </div>

              {/* Password */}
              <div className="space-y-2">
                <Label htmlFor="password">
                  Password <span className="text-destructive">*</span>
                </Label>
                <div className="relative">
                  <Input
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="••••••••"
                    {...register('password')}
                    onChange={(e) => {
                      register('password').onChange(e);
                      setPassword(e.target.value);
                    }}
                    aria-required="true"
                    aria-invalid={!!errors.password}
                    aria-describedby={errors.password ? 'password-error' : undefined}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    aria-label={showPassword ? 'Hide password' : 'Show password'}
                  >
                    {showPassword ? '🙈' : '👁️'}
                  </button>
                </div>
                <PasswordStrengthIndicator password={password} />
                <p className="text-xs text-muted-foreground">
                  Must be at least 8 characters with mix of letters, numbers, and symbols
                </p>
                {errors.password && (
                  <p id="password-error" role="alert" className="text-sm text-destructive">
                    {errors.password.message}
                  </p>
                )}
              </div>

              {/* Confirm Password */}
              <div className="space-y-2">
                <Label htmlFor="confirm_password">
                  Confirm Password <span className="text-destructive">*</span>
                </Label>
                <div className="relative">
                  <Input
                    id="confirm_password"
                    type={showConfirmPassword ? 'text' : 'password'}
                    placeholder="••••••••"
                    {...register('confirm_password')}
                    aria-required="true"
                    aria-invalid={!!errors.confirm_password}
                    aria-describedby={errors.confirm_password ? 'confirm-password-error' : undefined}
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
                  >
                    {showConfirmPassword ? '🙈' : '👁️'}
                  </button>
                </div>
                {errors.confirm_password && (
                  <p id="confirm-password-error" role="alert" className="text-sm text-destructive">
                    {errors.confirm_password.message}
                  </p>
                )}
              </div>
            </div>

            {/* Terms Checkbox */}
            <div className="flex items-start space-x-2">
              <Checkbox
                id="terms"
                checked={termsAccepted}
                onCheckedChange={(checked: boolean) => setValue('terms_accepted', checked)}
                aria-required="true"
                aria-invalid={!!errors.terms_accepted}
                aria-describedby={errors.terms_accepted ? 'terms-error' : undefined}
              />
              <div className="space-y-1">
                <label htmlFor="terms" className="text-sm leading-none cursor-pointer">
                  I agree to the{' '}
                  <Link href="/terms" target="_blank" className="text-primary hover:underline">
                    Terms of Service
                  </Link>{' '}
                  and{' '}
                  <Link href="/privacy" target="_blank" className="text-primary hover:underline">
                    Privacy Policy
                  </Link>
                </label>
                {errors.terms_accepted && (
                  <p id="terms-error" role="alert" className="text-sm text-destructive">
                    {errors.terms_accepted.message}
                  </p>
                )}
              </div>
            </div>

            {/* Submission Error */}
            {submissionError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  {errorCode === 'EMAIL_EXISTS'
                    ? 'This email is already registered. Please contact support or use a different email.'
                    : submissionError}
                </AlertDescription>
              </Alert>
            )}

            {/* Submit Button */}
            <Button type="submit" className="w-full" disabled={isSubmitting || !termsAccepted}>
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Accepting invitation...
                </>
              ) : (
                'Accept Invitation & Join Team'
              )}
            </Button>

            {/* Login Link */}
            <div className="text-center text-sm text-muted-foreground">
              Already have an account?{' '}
              <Link href="/login" className="text-primary hover:underline">
                Log in
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

export default function AcceptInvitePage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center bg-background p-4">
          <Card className="w-full max-w-lg">
            <CardContent className="py-8">
              <div
                role="status"
                aria-label="Loading page"
                className="flex flex-col items-center justify-center space-y-4"
              >
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
                <p className="text-sm text-muted-foreground">Loading...</p>
              </div>
            </CardContent>
          </Card>
        </div>
      }
    >
      <AcceptInviteContent />
    </Suspense>
  );
}
