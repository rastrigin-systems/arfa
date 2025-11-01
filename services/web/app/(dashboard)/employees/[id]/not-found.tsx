import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export default function EmployeeNotFound() {
  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>Employee Not Found</CardTitle>
          <CardDescription>
            The employee you are looking for does not exist or has been deleted.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div
            role="alert"
            aria-live="polite"
            className="text-sm text-muted-foreground mb-4"
          >
            Please check the employee ID and try again.
          </div>
          <Link href="/employees">
            <Button variant="default" className="w-full">
              Back to Employees
            </Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}
