import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { SettingsSidebar } from './SettingsSidebar';

// Mock next/navigation
vi.mock('next/navigation', () => ({
  usePathname: vi.fn(() => '/settings/organization'),
}));

describe('SettingsSidebar', () => {
  it('renders settings heading', () => {
    render(<SettingsSidebar isAdmin={true} />);
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('shows organization link for admins', () => {
    render(<SettingsSidebar isAdmin={true} />);
    expect(screen.getByText('Organization')).toBeInTheDocument();
  });

  it('hides organization link for non-admins', () => {
    render(<SettingsSidebar isAdmin={false} />);
    expect(screen.queryByText('Organization')).not.toBeInTheDocument();
  });

  it('always shows profile link', () => {
    render(<SettingsSidebar isAdmin={false} />);
    expect(screen.getByText('Profile')).toBeInTheDocument();
  });

  it('shows coming soon badge for disabled items', () => {
    render(<SettingsSidebar isAdmin={true} />);
    expect(screen.getAllByText('Coming Soon')).toHaveLength(3); // Billing, Security, Integrations
  });

  it('shows all navigation items for admin', () => {
    render(<SettingsSidebar isAdmin={true} />);
    expect(screen.getByText('Organization')).toBeInTheDocument();
    expect(screen.getByText('Profile')).toBeInTheDocument();
    expect(screen.getByText('Billing')).toBeInTheDocument();
    expect(screen.getByText('Security')).toBeInTheDocument();
    expect(screen.getByText('Integrations')).toBeInTheDocument();
  });
});
