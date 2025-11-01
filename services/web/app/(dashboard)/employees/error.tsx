'use client';

import { useEffect } from 'react';
import { Button } from '@/components/ui/button';

export default function EmployeesError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.error('Employees page error:', error);
  }, [error]);

  return (
    <div className="container mx-auto py-8">
      <div className="max-w-md mx-auto text-center space-y-4">
        <h2 className="text-2xl font-bold text-destructive">
          Something went wrong!
        </h2>
        <p className="text-muted-foreground">
          {error.message || 'An unexpected error occurred while loading employees.'}
        </p>
        <div className="flex gap-4 justify-center">
          <Button onClick={() => reset()}>Try again</Button>
          <Button variant="outline" onClick={() => (window.location.href = '/dashboard')}>
            Go to Dashboard
          </Button>
        </div>
      </div>
    </div>
  );
}
