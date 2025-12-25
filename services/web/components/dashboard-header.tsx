'use client';

import { LogOut, User } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/theme-toggle';
import { logoutAction } from '@/app/(dashboard)/actions';
import Link from 'next/link';

type Employee = {
  id: string;
  full_name: string;
  email: string;
  status: string;
};

export function DashboardHeader({ employee }: { employee: Employee }) {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-16 items-center justify-between px-6">
        <div className="flex items-center gap-6">
          <Link href="/dashboard" className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
              <span className="text-sm font-bold text-primary-foreground">U</span>
            </div>
            <h1 className="text-xl font-bold">Arfa Enterprise</h1>
          </Link>
        </div>

        <div className="flex items-center gap-4">
          {/* User info */}
          <div className="flex items-center gap-2 text-sm">
            <User className="h-4 w-4" />
            <span className="hidden font-medium sm:inline">{employee.full_name}</span>
            <span className="text-muted-foreground hidden md:inline">({employee.email})</span>
          </div>

          {/* Theme toggle */}
          <ThemeToggle />

          {/* Logout button */}
          <form action={logoutAction}>
            <Button variant="outline" size="sm" type="submit">
              <LogOut className="h-4 w-4" />
              <span className="hidden sm:inline ml-2">Logout</span>
            </Button>
          </form>
        </div>
      </div>
    </header>
  );
}
