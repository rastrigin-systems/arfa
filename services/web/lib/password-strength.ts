export type PasswordStrength = 'weak' | 'medium' | 'strong';

export interface PasswordRequirements {
  minLength: boolean;
  hasUppercase: boolean;
  hasNumber: boolean;
  hasSpecialChar: boolean;
}

export function checkPasswordRequirements(password: string): PasswordRequirements {
  return {
    minLength: password.length >= 8,
    hasUppercase: /[A-Z]/.test(password),
    hasNumber: /[0-9]/.test(password),
    hasSpecialChar: /[@$!%*?&]/.test(password),
  };
}

export function calculatePasswordStrength(password: string): PasswordStrength {
  const requirements = checkPasswordRequirements(password);
  const metCount = Object.values(requirements).filter(Boolean).length;

  if (metCount <= 1) return 'weak';
  if (metCount <= 3) return 'medium';
  return 'strong';
}

export function getStrengthColor(strength: PasswordStrength): string {
  switch (strength) {
    case 'weak':
      return 'bg-destructive';
    case 'medium':
      return 'bg-yellow-500';
    case 'strong':
      return 'bg-green-500';
  }
}

export function getStrengthWidth(strength: PasswordStrength): string {
  switch (strength) {
    case 'weak':
      return 'w-1/3';
    case 'medium':
      return 'w-2/3';
    case 'strong':
      return 'w-full';
  }
}
