'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  Users,
  UsersRound,
  Bot,
  History,
  Settings,
  ChevronRight,
  Building2,
  FileStack,
  Server,
  ClipboardList,
  Shield,
  Bug,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface NavItemProps {
  href: string;
  icon: React.ReactNode;
  label: string;
  active?: boolean;
  badge?: number;
}

function NavItem({ href, icon, label, active, badge }: NavItemProps) {
  return (
    <Link
      href={href}
      className={cn(
        'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
        active
          ? 'bg-primary text-primary-foreground'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
      )}
    >
      {icon}
      <span className="flex-1">{label}</span>
      {badge !== undefined && (
        <span
          className={cn(
            'rounded-full px-2 py-0.5 text-xs',
            active ? 'bg-primary-foreground/20 text-primary-foreground' : 'bg-muted-foreground/20'
          )}
        >
          {badge}
        </span>
      )}
      {active && <ChevronRight className="h-4 w-4" />}
    </Link>
  );
}

interface SidebarProps {
  teamCount?: number;
  employeeCount?: number;
}

export function Sidebar({ teamCount, employeeCount }: SidebarProps) {
  const pathname = usePathname();

  const navItems = [
    {
      href: '/dashboard',
      icon: <LayoutDashboard className="h-4 w-4" />,
      label: 'Dashboard',
    },
    {
      href: '/teams',
      icon: <UsersRound className="h-4 w-4" />,
      label: 'Teams',
      badge: teamCount,
    },
    {
      href: '/employees',
      icon: <Users className="h-4 w-4" />,
      label: 'Employees',
      badge: employeeCount,
    },
    {
      href: '/roles',
      icon: <Shield className="h-4 w-4" />,
      label: 'Roles',
    },
    {
      href: '/agents',
      icon: <Bot className="h-4 w-4" />,
      label: 'Agents',
    },
    {
      href: '/configs',
      icon: <FileStack className="h-4 w-4" />,
      label: 'Configs',
    },
    {
      href: '/mcp',
      icon: <Server className="h-4 w-4" />,
      label: 'MCP',
    },
    {
      href: '/requests',
      icon: <ClipboardList className="h-4 w-4" />,
      label: 'Requests',
    },
    {
      href: '/logs',
      icon: <History className="h-4 w-4" />,
      label: 'Activity Logs',
    },
    {
      href: '/debug',
      icon: <Bug className="h-4 w-4" />,
      label: 'Debug Logs',
    },
  ];

  const bottomNavItems = [
    {
      href: '/settings',
      icon: <Settings className="h-4 w-4" />,
      label: 'Settings',
    },
  ];

  return (
    <aside className="fixed left-0 top-16 z-30 h-[calc(100vh-4rem)] w-64 border-r bg-background">
      <div className="flex h-full flex-col gap-2 p-4">
        {/* Organization Header */}
        <div className="mb-4 flex items-center gap-2 rounded-lg bg-muted p-3">
          <Building2 className="h-5 w-5 text-muted-foreground" />
          <div className="flex-1 truncate">
            <p className="text-sm font-medium">Organization</p>
            <p className="text-xs text-muted-foreground truncate">Manage your workspace</p>
          </div>
        </div>

        {/* Main Navigation */}
        <nav className="flex-1 space-y-1">
          {navItems.map((item) => (
            <NavItem
              key={item.href}
              href={item.href}
              icon={item.icon}
              label={item.label}
              active={pathname === item.href || pathname.startsWith(`${item.href}/`)}
              badge={item.badge}
            />
          ))}
        </nav>

        {/* Bottom Navigation */}
        <div className="border-t pt-4 space-y-1">
          {bottomNavItems.map((item) => (
            <NavItem
              key={item.href}
              href={item.href}
              icon={item.icon}
              label={item.label}
              active={pathname === item.href || pathname.startsWith(`${item.href}/`)}
            />
          ))}
        </div>
      </div>
    </aside>
  );
}
