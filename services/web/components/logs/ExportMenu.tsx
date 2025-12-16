'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Download, Loader2 } from 'lucide-react';

interface ExportMenuProps {
  filters: {
    session_id?: string;
    employee_id?: string;
    agent_id?: string;
    event_type?: string;
    event_category?: string;
    start_date?: string;
    end_date?: string;
    search?: string;
  };
}

export function ExportMenu({ filters }: ExportMenuProps) {
  const { toast } = useToast();
  const [format, setFormat] = useState<'json' | 'csv'>('json');
  const [isExporting, setIsExporting] = useState(false);

  const handleExport = async () => {
    setIsExporting(true);

    try {
      const params = new URLSearchParams();
      params.append('format', format);
      if (filters.session_id) params.append('session_id', filters.session_id);
      if (filters.employee_id) params.append('employee_id', filters.employee_id);
      if (filters.agent_id) params.append('agent_id', filters.agent_id);
      if (filters.event_type) params.append('event_type', filters.event_type);
      if (filters.event_category) params.append('event_category', filters.event_category);
      if (filters.start_date) params.append('start_date', filters.start_date);
      if (filters.end_date) params.append('end_date', filters.end_date);
      if (filters.search) params.append('search', filters.search);

      const response = await fetch(`/api/logs/export?${params.toString()}`);

      if (!response.ok) {
        throw new Error('Failed to export logs');
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `activity-logs-${new Date().toISOString()}.${format}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch {
      toast({
        title: 'Export failed',
        description: 'Failed to export logs. Please try again.',
        variant: 'destructive',
      });
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <Select value={format} onValueChange={(value: 'json' | 'csv') => setFormat(value)}>
        <SelectTrigger className="w-24">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="json">JSON</SelectItem>
          <SelectItem value="csv">CSV</SelectItem>
        </SelectContent>
      </Select>

      <Button onClick={handleExport} disabled={isExporting}>
        {isExporting ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Exporting...
          </>
        ) : (
          <>
            <Download className="mr-2 h-4 w-4" />
            Export
          </>
        )}
      </Button>
    </div>
  );
}
