'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ChevronRight, LucideIcon } from 'lucide-react';

type ConfigurationLevelCardProps = {
  icon: LucideIcon;
  title: string;
  description: string;
  count?: number;
  countLabel?: string;
  actionLabel: string;
  onAction: () => void;
  variant?: 'primary' | 'secondary' | 'tertiary';
};

export function ConfigurationLevelCard({
  icon: Icon,
  title,
  description,
  count,
  countLabel,
  actionLabel,
  onAction,
  variant = 'primary',
}: ConfigurationLevelCardProps) {
  const variantStyles = {
    primary: 'border-primary bg-primary/5',
    secondary: 'border-blue-500 bg-blue-50 dark:bg-blue-950',
    tertiary: 'border-green-500 bg-green-50 dark:bg-green-950',
  };

  const iconStyles = {
    primary: 'bg-primary text-primary-foreground',
    secondary: 'bg-blue-500 text-white',
    tertiary: 'bg-green-500 text-white',
  };

  return (
    <Card className={variantStyles[variant]}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={`flex h-10 w-10 items-center justify-center rounded-full ${iconStyles[variant]}`}>
              <Icon className="h-5 w-5" />
            </div>
            <div>
              <CardTitle className="text-xl">{title}</CardTitle>
              <CardDescription>{description}</CardDescription>
            </div>
          </div>
          {count !== undefined && (
            <Badge variant="secondary">
              {count} {countLabel}
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <Button onClick={onAction} className="w-full justify-between">
          {actionLabel}
          <ChevronRight className="h-4 w-4" />
        </Button>
      </CardContent>
    </Card>
  );
}
