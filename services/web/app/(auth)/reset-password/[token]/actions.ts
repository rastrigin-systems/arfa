'use server';

import { apiClient } from '@/lib/api/client';
import { z } from 'zod';

const resetPasswordSchema = z
  .object({
    token: z.string().min(1, 'Token is required'),
    new_password: z
      .string()
      .min(8, 'Password must be at least 8 characters')
      .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
      .regex(/[0-9]/, 'Password must contain at least one number')
      .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character'),
    confirm_password: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: 'Passwords do not match',
    path: ['confirm_password'],
  });

export type ResetPasswordFormState = {
  errors?: {
    new_password?: string[];
    confirm_password?: string[];
    _form?: string[];
  };
  success?: boolean;
  message?: string;
};

export async function resetPasswordAction(
  prevState: ResetPasswordFormState,
  formData: FormData
): Promise<ResetPasswordFormState> {
  // Validate form data
  const validatedFields = resetPasswordSchema.safeParse({
    token: formData.get('token'),
    new_password: formData.get('new_password'),
    confirm_password: formData.get('confirm_password'),
  });

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  const { token, new_password } = validatedFields.data;

  try {
    // Call reset-password API
    const { data, error, response } = await apiClient.POST('/auth/reset-password', {
      body: {
        token,
        new_password,
      },
    });

    if (error || !response.ok) {
      // Handle specific error cases
      const errorMessage = error?.message || 'An error occurred. Please try again.';

      return {
        errors: {
          _form: [errorMessage],
        },
      };
    }

    // Success
    return {
      success: true,
      message: data?.message || 'Password reset successful',
    };
  } catch (err) {
    console.error('Reset password error:', err);
    return {
      errors: {
        _form: ['An unexpected error occurred. Please try again.'],
      },
    };
  }
}

// Verify token validity
export async function verifyResetToken(token: string): Promise<{
  valid: boolean;
  error?: string;
}> {
  try {
    const { data, error, response } = await apiClient.GET('/auth/verify-reset-token', {
      params: {
        query: {
          token,
        },
      },
    });

    if (error || !response.ok) {
      return {
        valid: false,
        error: error?.message || 'Invalid or expired token',
      };
    }

    return {
      valid: data?.valid ?? false,
    };
  } catch (err) {
    console.error('Verify token error:', err);
    return {
      valid: false,
      error: 'An unexpected error occurred',
    };
  }
}
