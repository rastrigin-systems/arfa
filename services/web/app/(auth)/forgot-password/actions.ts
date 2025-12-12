'use server';

import { apiClient } from '@/lib/api/client';
import { z } from 'zod';

const forgotPasswordSchema = z.object({
  email: z.string().email('Invalid email address'),
});

export type ForgotPasswordFormState = {
  errors?: {
    email?: string[];
    _form?: string[];
  };
  success?: boolean;
  message?: string;
};

export async function forgotPasswordAction(
  prevState: ForgotPasswordFormState,
  formData: FormData
): Promise<ForgotPasswordFormState> {
  // Validate form data
  const validatedFields = forgotPasswordSchema.safeParse({
    email: formData.get('email'),
  });

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  const { email } = validatedFields.data;

  try {
    // Call forgot-password API
    const { data, error, response } = await apiClient.POST('/auth/forgot-password', {
      body: {
        email,
      },
    });

    if (error || !response.ok) {
      // Handle rate limiting (429)
      if (response.status === 429) {
        return {
          errors: {
            _form: ['You have requested the maximum number of reset links per hour. Please try again later.'],
          },
        };
      }

      // Extract error message from API response
      const errorMessage = error?.error || error?.message || 'An error occurred. Please try again.';
      return {
        errors: {
          _form: [errorMessage],
        },
      };
    }

    // Always return success with generic message (security - no email enumeration)
    return {
      success: true,
      message: data?.message || 'Password reset link sent to your email',
    };
  } catch (err) {
    console.error('Forgot password error:', err);
    return {
      errors: {
        _form: ['An unexpected error occurred. Please try again.'],
      },
    };
  }
}
