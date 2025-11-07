'use client';

import { CheckCircle2, XCircle } from 'lucide-react';
import {
  calculatePasswordStrength,
  checkPasswordRequirements,
  getStrengthColor,
  getStrengthWidth,
  type PasswordStrength,
} from '@/lib/password-strength';

interface PasswordStrengthIndicatorProps {
  password: string;
}

export function PasswordStrengthIndicator({ password }: PasswordStrengthIndicatorProps) {
  if (!password) return null;

  const strength = calculatePasswordStrength(password);
  const requirements = checkPasswordRequirements(password);
  const strengthColor = getStrengthColor(strength);
  const strengthWidth = getStrengthWidth(strength);

  const strengthText: Record<PasswordStrength, string> = {
    weak: 'Weak',
    medium: 'Medium',
    strong: 'Strong',
  };

  return (
    <div className="space-y-3">
      {/* Strength indicator bar */}
      <div className="space-y-1">
        <div className="flex items-center justify-between text-sm">
          <span className="font-medium">Password Strength:</span>
          <span className={strengthColor === 'bg-destructive' ? 'text-destructive' : strengthColor === 'bg-yellow-500' ? 'text-yellow-600' : 'text-green-600'}>
            {strengthText[strength]}
          </span>
        </div>
        <div className="h-2 w-full bg-muted rounded-full overflow-hidden">
          <div
            className={`h-full transition-all duration-300 ${strengthColor} ${strengthWidth}`}
            role="progressbar"
            aria-valuenow={strength === 'weak' ? 33 : strength === 'medium' ? 66 : 100}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label={`Password strength: ${strengthText[strength]}`}
          />
        </div>
      </div>

      {/* Requirements checklist */}
      <div className="space-y-1">
        <p className="text-sm font-medium">Requirements:</p>
        <div className="space-y-1">
          <RequirementItem
            met={requirements.minLength}
            text="At least 8 characters"
          />
          <RequirementItem
            met={requirements.hasUppercase}
            text="One uppercase letter"
          />
          <RequirementItem
            met={requirements.hasNumber}
            text="One number"
          />
          <RequirementItem
            met={requirements.hasSpecialChar}
            text="One special character"
          />
        </div>
      </div>
    </div>
  );
}

interface RequirementItemProps {
  met: boolean;
  text: string;
}

function RequirementItem({ met, text }: RequirementItemProps) {
  return (
    <div className="flex items-center gap-2 text-sm">
      {met ? (
        <CheckCircle2 className="h-4 w-4 text-green-600" aria-hidden="true" />
      ) : (
        <XCircle className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
      )}
      <span className={met ? 'text-foreground' : 'text-muted-foreground'}>
        {text}
      </span>
    </div>
  );
}
