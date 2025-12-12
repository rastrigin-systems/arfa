'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ArrowDown, Building2, Users, User } from 'lucide-react';

export function ConfigurationHierarchy() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>How Configuration Works</CardTitle>
        <CardDescription>
          Agent settings cascade from organization defaults down to individual employees
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col items-center gap-4 py-4">
          {/* Organization Level */}
          <div className="flex w-full max-w-md flex-col items-center">
            <div className="flex w-full items-center gap-3 rounded-lg border-2 border-primary bg-primary/5 p-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground">
                <Building2 className="h-5 w-5" />
              </div>
              <div className="flex-1">
                <div className="font-semibold">Organization Defaults</div>
                <div className="text-sm text-muted-foreground">Base settings for everyone</div>
              </div>
            </div>
            <ArrowDown className="my-2 h-6 w-6 text-muted-foreground" />
          </div>

          {/* Team Level */}
          <div className="flex w-full max-w-md flex-col items-center">
            <div className="flex w-full items-center gap-3 rounded-lg border-2 border-blue-500 bg-blue-50 p-4 dark:bg-blue-950">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-500 text-white">
                <Users className="h-5 w-5" />
              </div>
              <div className="flex-1">
                <div className="font-semibold">Team Overrides</div>
                <div className="text-sm text-muted-foreground">Optional team customizations</div>
              </div>
            </div>
            <ArrowDown className="my-2 h-6 w-6 text-muted-foreground" />
          </div>

          {/* Employee Level */}
          <div className="flex w-full max-w-md flex-col items-center">
            <div className="flex w-full items-center gap-3 rounded-lg border-2 border-green-500 bg-green-50 p-4 dark:bg-green-950">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-green-500 text-white">
                <User className="h-5 w-5" />
              </div>
              <div className="flex-1">
                <div className="font-semibold">Employee Overrides</div>
                <div className="text-sm text-muted-foreground">Individual customizations</div>
              </div>
            </div>
          </div>

          {/* Explanation */}
          <div className="mt-4 rounded-md bg-muted/50 p-4 text-sm text-muted-foreground">
            <p>
              <strong>Example:</strong> If you enable Claude at the organization level with GPT-4, all employees get
              Claude with GPT-4. Teams can override to use GPT-3.5 instead. Individual employees can override to use
              a different model or disable Claude entirely.
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
