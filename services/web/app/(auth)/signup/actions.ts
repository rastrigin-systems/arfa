'use server';

import { redirect } from 'next/navigation';
import { apiClient } from '@/lib/api/client';
import { getErrorMessage } from '@/lib/api/errors';
import { setServerToken } from '@/lib/auth';
import { z } from 'zod';

const signupSchema = z
  .object({
    full_name: z
      .string()
      .min(2, 'Full name must be at least 2 characters')
      .max(100, 'Full name must not exceed 100 characters'),
    email: z.string().email('Invalid email address'),
    org_name: z
      .string()
      .min(2, 'Organization name must be at least 2 characters')
      .max(100, 'Organization name must not exceed 100 characters'),
    org_slug: z
      .string()
      .min(3, 'Slug must be 3-50 characters')
      .max(50, 'Slug must be 3-50 characters')
      .regex(/^[a-z][a-z0-9-]*$/, 'Slug can only contain lowercase letters, numbers, and hyphens')
      .refine((val) => /^[a-z]/.test(val), {
        message: 'Slug must start with a letter',
      }),
    password: z
      .string()
      .min(8, 'Password must be at least 8 characters')
      .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
      .regex(/[0-9]/, 'Password must contain at least one number')
      .regex(/[@$!%*?&]/, 'Password must contain at least one special character (@$!%*?&)'),
    confirm_password: z.string(),
  })
  .refine((data) => data.password === data.confirm_password, {
    message: 'Passwords do not match',
    path: ['confirm_password'],
  });

export type SignupFormState = {
  errors?: {
    full_name?: string[];
    email?: string[];
    org_name?: string[];
    org_slug?: string[];
    password?: string[];
    confirm_password?: string[];
    _form?: string[];
  };
  success?: boolean;
};

export async function signupAction(
  prevState: SignupFormState,
  formData: FormData
): Promise<SignupFormState> {
  // Validate form data
  const validatedFields = signupSchema.safeParse({
    full_name: formData.get('full_name'),
    email: formData.get('email'),
    org_name: formData.get('org_name'),
    org_slug: formData.get('org_slug'),
    password: formData.get('password'),
    confirm_password: formData.get('confirm_password'),
  });

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  const { full_name, email, org_name, org_slug, password } = validatedFields.data;

  try {
    // Call register API
    const { data, error } = await apiClient.POST('/auth/register', {
      body: {
        full_name,
        email,
        org_name,
        org_slug,
        password,
      },
    });

    if (error) {
      const errorMessage = getErrorMessage(error, 'Registration failed. Please try again.');
      return {
        errors: {
          _form: [errorMessage],
        },
      };
    }

    if (!data) {
      return {
        errors: {
          _form: ['Registration failed. Please try again.'],
        },
      };
    }

    // Store token in httpOnly cookie
    await setServerToken(data.token);
  } catch (err) {
    console.error('Signup error:', err);
    return {
      errors: {
        _form: ['An unexpected error occurred. Please try again.'],
      },
    };
  }

  // Redirect outside try/catch so Next.js redirect error propagates correctly
  redirect('/dashboard');
}

export async function checkSlugAvailability(slug: string): Promise<{ available: boolean }> {
  try {
    const { data, error } = await apiClient.GET('/auth/check-slug', {
      params: {
        query: { slug },
      },
    });

    if (error || !data) {
      // Default to unavailable on error
      return { available: false };
    }

    return { available: data.available };
  } catch (err) {
    console.error('Slug availability check error:', err);
    // Default to unavailable on error
    return { available: false };
  }
}
