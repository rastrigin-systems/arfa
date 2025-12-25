import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import SignupPage from './page';

// Mock react-dom hooks (required for Next.js 14 form actions)
vi.mock('react-dom', async () => {
  const actual = await vi.importActual('react-dom');
  return {
    ...actual,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    useFormState: (action: any, initialState: any) => {
      // Return state and a mock action function
      return [initialState, action];
    },
    useFormStatus: () => ({
      pending: false,
      data: null,
      method: null,
      action: null,
    }),
  };
});

// Mock server actions
vi.mock('./actions', () => ({
  signupAction: vi.fn(),
  checkSlugAvailability: vi.fn().mockResolvedValue({ available: true }),
}));

describe('SignupPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Form Rendering', () => {
    it('should render signup page with title', () => {
      render(<SignupPage />);

      expect(screen.getByText('Arfa')).toBeInTheDocument();
      expect(screen.getByText(/Create Your Organization Account/i)).toBeInTheDocument();
    });

    it('should render all form fields', () => {
      render(<SignupPage />);

      expect(screen.getByLabelText(/Full Name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^Email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Organization Name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Organization Slug/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^Password$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/Confirm Password/i)).toBeInTheDocument();
    });

    it('should render submit button', () => {
      render(<SignupPage />);

      const submitButton = screen.getByRole('button', { name: /Create Account/i });
      expect(submitButton).toBeInTheDocument();
      expect(submitButton).toHaveAttribute('type', 'submit');
    });

    it('should render link to login page', () => {
      render(<SignupPage />);

      const loginLink = screen.getByRole('link', { name: /Sign in/i });
      expect(loginLink).toBeInTheDocument();
      expect(loginLink).toHaveAttribute('href', '/login');
    });

    it('should have all required fields marked as required', () => {
      render(<SignupPage />);

      expect(screen.getByLabelText(/Full Name/i)).toHaveAttribute('required');
      expect(screen.getByLabelText(/^Email/i)).toHaveAttribute('required');
      expect(screen.getByLabelText(/Organization Name/i)).toHaveAttribute('required');
      expect(screen.getByLabelText(/Organization Slug/i)).toHaveAttribute('required');
      expect(screen.getByLabelText(/^Password$/i)).toHaveAttribute('required');
      expect(screen.getByLabelText(/Confirm Password/i)).toHaveAttribute('required');
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA attributes on form fields', () => {
      render(<SignupPage />);

      const nameInput = screen.getByLabelText(/Full Name/i);
      expect(nameInput).toHaveAttribute('aria-required', 'true');

      const emailInput = screen.getByLabelText(/^Email/i);
      expect(emailInput).toHaveAttribute('aria-required', 'true');

      const passwordInput = screen.getByLabelText(/^Password$/i);
      expect(passwordInput).toHaveAttribute('aria-required', 'true');
    });
  });

  describe('Responsive Design', () => {
    it('should have responsive container classes', () => {
      const { container } = render(<SignupPage />);

      const mainDiv = container.querySelector('[class*="min-h-screen"]');
      expect(mainDiv).toHaveClass('flex', 'items-center', 'justify-center');
    });

    it('should have max-width constraint on card', () => {
      const { container } = render(<SignupPage />);

      const card = container.querySelector('[class*="max-w"]');
      expect(card).toBeInTheDocument();
    });
  });
});
