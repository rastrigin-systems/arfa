'use server';

import { redirect } from 'next/navigation';
import { apiClient } from '@/lib/api/client';
import { setServerToken } from '@/lib/auth';
import { z } from 'zod';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
});

export type LoginFormState = {
  errors?: {
    email?: string[];
    password?: string[];
    _form?: string[];
  };
  success?: boolean;
};

export async function loginAction(
  prevState: LoginFormState,
  formData: FormData
): Promise<LoginFormState> {
  // Validate form data
  const validatedFields = loginSchema.safeParse({
    email: formData.get('email'),
    password: formData.get('password'),
  });

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  const { email, password } = validatedFields.data;

  try {
    // Call login API
    const { data, error, response } = await apiClient.POST('/auth/login', {
      body: {
        email,
        password,
      },
    });

    if (error || !response.ok) {
      return {
        errors: {
          _form: [error?.message || 'Invalid credentials. Please try again.'],
        },
      };
    }

    if (!data) {
      return {
        errors: {
          _form: ['Login failed. Please try again.'],
        },
      };
    }

    // Store token in httpOnly cookie
    await setServerToken(data.token);

    return { success: true };
  } catch (err) {
    console.error('Login error:', err);
    return {
      errors: {
        _form: ['An unexpected error occurred. Please try again.'],
      },
    };
  }
}

// This function is called after successful login to redirect
export async function redirectToDashboard() {
  redirect('/dashboard');
}
