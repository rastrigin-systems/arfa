'use client';

import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface LogFiltersProps {
  filters: {
    employee_id?: string;
    agent_id?: string;
    event_type?: string;
    start_date?: string;
    end_date?: string;
    search?: string;
  };
  onChange: (filters: Record<string, string | undefined>) => void;
}

export function LogFilters({ filters, onChange }: LogFiltersProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* Search */}
      <div className="space-y-2">
        <Label htmlFor="search">Search Logs</Label>
        <Input
          id="search"
          type="text"
          placeholder="Search logs..."
          value={filters.search || ''}
          onChange={(e) => onChange({ search: e.target.value })}
        />
      </div>

      {/* Date Range */}
      <div className="space-y-2">
        <Label htmlFor="date-range">Date Range</Label>
        <div className="flex gap-2">
          <Input
            id="start-date"
            type="date"
            value={filters.start_date || ''}
            onChange={(e) => onChange({ start_date: e.target.value })}
            aria-label="Start date"
          />
          <Input
            id="end-date"
            type="date"
            value={filters.end_date || ''}
            onChange={(e) => onChange({ end_date: e.target.value })}
            aria-label="End date"
          />
        </div>
      </div>

      {/* Employee */}
      <div className="space-y-2">
        <Label htmlFor="employee">Employee</Label>
        <Select
          value={filters.employee_id || ''}
          onValueChange={(value) => onChange({ employee_id: value || undefined })}
        >
          <SelectTrigger id="employee">
            <SelectValue placeholder="All employees" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All employees</SelectItem>
            {/* TODO: Load from API */}
          </SelectContent>
        </Select>
      </div>

      {/* Agent */}
      <div className="space-y-2">
        <Label htmlFor="agent">Agent</Label>
        <Select
          value={filters.agent_id || ''}
          onValueChange={(value) => onChange({ agent_id: value || undefined })}
        >
          <SelectTrigger id="agent">
            <SelectValue placeholder="All agents" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All agents</SelectItem>
            {/* TODO: Load from API */}
          </SelectContent>
        </Select>
      </div>

      {/* Event Type */}
      <div className="space-y-2">
        <Label htmlFor="event-type">Event Type</Label>
        <Select
          value={filters.event_type || ''}
          onValueChange={(value) => onChange({ event_type: value || undefined })}
        >
          <SelectTrigger id="event-type">
            <SelectValue placeholder="All types" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All types</SelectItem>
            <SelectItem value="input">Input</SelectItem>
            <SelectItem value="output">Output</SelectItem>
            <SelectItem value="error">Error</SelectItem>
            <SelectItem value="session_start">Session Start</SelectItem>
            <SelectItem value="session_end">Session End</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  );
}
