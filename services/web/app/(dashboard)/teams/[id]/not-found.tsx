import Link from 'next/link';
import { UsersRound } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function TeamNotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
      <div className="flex h-16 w-16 items-center justify-center rounded-full bg-muted">
        <UsersRound className="h-8 w-8 text-muted-foreground" />
      </div>
      <h2 className="mt-6 text-2xl font-bold">Team not found</h2>
      <p className="mt-2 text-muted-foreground max-w-md">
        The team you're looking for doesn't exist or has been deleted.
      </p>
      <Link href="/teams" className="mt-6">
        <Button>Back to Teams</Button>
      </Link>
    </div>
  );
}
