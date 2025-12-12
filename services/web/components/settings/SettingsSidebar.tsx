'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Building2,
  User,
  CreditCard,
  Shield,
  Puzzle,
  ChevronRight,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';

interface SettingsSidebarProps {
  isAdmin: boolean;
}

interface NavItemProps {
  href: string;
  icon: React.ReactNode;
  label: string;
  active?: boolean;
  disabled?: boolean;
  comingSoon?: boolean;
}

function NavItem({ href, icon, label, active, disabled, comingSoon }: NavItemProps) {
  const content = (
    <div
      className={cn(
        'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
        active
          ? 'bg-primary text-primary-foreground'
          : disabled
            ? 'text-muted-foreground cursor-not-allowed'
            : 'text-muted-foreground hover:bg-muted hover:text-foreground'
      )}
    >
      {icon}
      <span className="flex-1">{label}</span>
      {comingSoon && (
        <Badge variant="secondary" className="text-xs">
          Coming Soon
        </Badge>
      )}
      {active && <ChevronRight className="h-4 w-4" />}
    </div>
  );

  if (disabled) {
    return content;
  }

  return <Link href={href}>{content}</Link>;
}

export function SettingsSidebar({ isAdmin }: SettingsSidebarProps) {
  const pathname = usePathname();

  const navItems = [
    // Organization - only visible to admins
    ...(isAdmin
      ? [
          {
            href: '/settings/organization',
            icon: <Building2 className="h-4 w-4" />,
            label: 'Organization',
            disabled: false,
            comingSoon: false,
          },
        ]
      : []),
    {
      href: '/settings/profile',
      icon: <User className="h-4 w-4" />,
      label: 'Profile',
      disabled: false,
      comingSoon: false,
    },
    {
      href: '/settings/billing',
      icon: <CreditCard className="h-4 w-4" />,
      label: 'Billing',
      disabled: true,
      comingSoon: true,
    },
    {
      href: '/settings/security',
      icon: <Shield className="h-4 w-4" />,
      label: 'Security',
      disabled: true,
      comingSoon: true,
    },
    {
      href: '/settings/integrations',
      icon: <Puzzle className="h-4 w-4" />,
      label: 'Integrations',
      disabled: true,
      comingSoon: true,
    },
  ];

  return (
    <aside className="w-60 shrink-0">
      <div className="sticky top-6">
        <h2 className="mb-4 text-lg font-semibold">Settings</h2>
        <nav className="space-y-1" aria-label="Settings navigation">
          {navItems.map((item) => (
            <NavItem
              key={item.href}
              href={item.href}
              icon={item.icon}
              label={item.label}
              active={pathname === item.href}
              disabled={item.disabled}
              comingSoon={item.comingSoon}
            />
          ))}
        </nav>
      </div>
    </aside>
  );
}
