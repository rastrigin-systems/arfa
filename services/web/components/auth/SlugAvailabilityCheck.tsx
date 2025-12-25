'use client';

import { useEffect, useState } from 'react';
import { CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import { checkSlugAvailability } from '@/app/(auth)/signup/actions';

interface SlugAvailabilityCheckProps {
  slug: string;
  onAvailabilityChange: (available: boolean) => void;
}

type CheckState = 'idle' | 'checking' | 'available' | 'unavailable';

export function SlugAvailabilityCheck({ slug, onAvailabilityChange }: SlugAvailabilityCheckProps) {
  const [checkState, setCheckState] = useState<CheckState>('idle');

  useEffect(() => {
    // Don't check if slug is empty or invalid format
    if (!slug || slug.length < 3) {
      setCheckState('idle');
      onAvailabilityChange(false);
      return;
    }

    // Validate format before checking
    const isValidFormat = /^[a-z][a-z0-9-]*$/.test(slug);
    if (!isValidFormat) {
      setCheckState('idle');
      onAvailabilityChange(false);
      return;
    }

    setCheckState('checking');

    // Debounce the API call
    const timeoutId = setTimeout(async () => {
      try {
        const result = await checkSlugAvailability(slug);
        setCheckState(result.available ? 'available' : 'unavailable');
        onAvailabilityChange(result.available);
      } catch (error) {
        console.error('Error checking slug availability:', error);
        setCheckState('unavailable');
        onAvailabilityChange(false);
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [slug, onAvailabilityChange]);

  if (checkState === 'idle') return null;

  return (
    <div
      className="flex items-center gap-2 text-sm mt-1"
      role="status"
      aria-live="polite"
    >
      {checkState === 'checking' && (
        <>
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" aria-hidden="true" />
          <span className="text-muted-foreground">Checking availability...</span>
        </>
      )}

      {checkState === 'available' && (
        <>
          <CheckCircle2 className="h-4 w-4 text-green-600" aria-hidden="true" />
          <span className="text-green-600">{slug}.arfa.com is available</span>
        </>
      )}

      {checkState === 'unavailable' && (
        <>
          <XCircle className="h-4 w-4 text-destructive" aria-hidden="true" />
          <span className="text-destructive">This slug is already taken</span>
        </>
      )}
    </div>
  );
}
