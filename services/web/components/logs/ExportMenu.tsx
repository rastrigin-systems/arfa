'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Download, Loader2 } from 'lucide-react';
import { apiClient } from '@/lib/api/client';

interface ExportMenuProps {
  filters: Record<string, string | undefined>;
}

export function ExportMenu({ filters }: ExportMenuProps) {
  const [format, setFormat] = useState<'json' | 'csv'>('json');
  const [isExporting, setIsExporting] = useState(false);

  const handleExport = async () => {
    setIsExporting(true);

    try {
      const { data, error } = await apiClient.GET('/logs/export', {
        params: {
          query: {
            format,
            ...filters,
          },
        },
      });

      if (error) {
        throw new Error('Failed to export logs');
      }

      // Create a blob and download
      const blob = new Blob([data as BlobPart], {
        type: format === 'json' ? 'application/json' : 'text/csv',
      });

      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `activity-logs-${new Date().toISOString()}.${format}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Export failed:', err);
      alert('Failed to export logs. Please try again.');
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
