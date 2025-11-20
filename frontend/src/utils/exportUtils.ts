/**
 * Utility functions for exporting data to CSV and JSON formats
 */

/**
 * Convert array of objects to CSV string
 */
export function arrayToCSV<T extends Record<string, any>>(
  data: T[],
  headers?: string[]
): string {
  if (!data || data.length === 0) {
    return '';
  }

  // Use provided headers or extract from first object
  const columnHeaders = headers || Object.keys(data[0]);

  // Create CSV header row
  const csvHeader = columnHeaders.join(',');

  // Create CSV rows
  const csvRows = data.map((row) => {
    return columnHeaders
      .map((header) => {
        const value = row[header];
        // Handle null/undefined
        if (value === null || value === undefined) {
          return '';
        }
        // Handle strings with commas or quotes
        const stringValue = String(value);
        if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) {
          return `"${stringValue.replace(/"/g, '""')}"`;
        }
        return stringValue;
      })
      .join(',');
  });

  return [csvHeader, ...csvRows].join('\n');
}

/**
 * Download data as CSV file
 */
export function downloadCSV<T extends Record<string, any>>(
  data: T[],
  filename: string,
  headers?: string[]
): void {
  const csv = arrayToCSV(data, headers);
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  downloadBlob(blob, `${filename}.csv`);
}

/**
 * Download data as JSON file
 */
export function downloadJSON<T>(data: T, filename: string): void {
  const json = JSON.stringify(data, null, 2);
  const blob = new Blob([json], { type: 'application/json;charset=utf-8;' });
  downloadBlob(blob, `${filename}.json`);
}

/**
 * Helper to trigger download of a blob
 */
function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/**
 * Format host data for export
 */
export function formatHostsForExport(hosts: any[]): any[] {
  return hosts.map((host) => ({
    ID: host.id,
    Hostname: host.hostname,
    IP: host.ip,
    OS: host.os,
    Status: host.agent_status || 'offline',
    Group: host.group || 'default',
    Tags: Array.isArray(host.tags) ? host.tags.join('; ') : host.tags || '',
    'SSH User': host.ssh_user || '',
    'SSH Port': host.ssh_port || 22,
    Description: host.description || '',
    'Created At': host.created_at ? new Date(host.created_at).toISOString() : '',
    'Updated At': host.updated_at ? new Date(host.updated_at).toISOString() : '',
  }));
}

/**
 * Format metrics data for export
 */
export function formatMetricsForExport(metrics: any[]): any[] {
  return metrics.map((metric) => ({
    Timestamp: metric.timestamp ? new Date(metric.timestamp).toISOString() : '',
    'Host ID': metric.host_id,
    Hostname: metric.hostname || '',
    Name: metric.name,
    Value: metric.value,
    Unit: metric.unit || '',
    Tags: typeof metric.tags === 'object' ? JSON.stringify(metric.tags) : metric.tags || '',
  }));
}

/**
 * Format logs data for export
 */
export function formatLogsForExport(logs: any[]): any[] {
  return logs.map((log) => ({
    Timestamp: log.timestamp ? new Date(log.timestamp).toISOString() : '',
    'Host ID': log.host_id,
    Hostname: log.hostname || '',
    Level: log.level || 'info',
    Message: log.message || '',
    Source: log.source || '',
    Tags: typeof log.tags === 'object' ? JSON.stringify(log.tags) : log.tags || '',
  }));
}
