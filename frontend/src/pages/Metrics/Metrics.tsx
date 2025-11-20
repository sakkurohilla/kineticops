import React, { useState, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import MetricsChart from '../../components/metrics/MetricsChart';
import TimeRangeSelector, { TimeRange } from '../../components/metrics/TimeRangeSelector';
import { useMetrics } from '../../hooks/useMetrics';
import Button from '../../components/common/Button';
import { RefreshCw, Download, Server } from 'lucide-react';
import hostService from '../../services/api/hostService';
import { Host } from '../../types';

const Metrics: React.FC = () => {
  const [timeRange, setTimeRange] = useState<TimeRange>('24h');
  const [customStart, setCustomStart] = useState<string>();
  const [customEnd, setCustomEnd] = useState<string>();
  const [selectedHostId, setSelectedHostId] = useState<number | undefined>();
  const [hosts, setHosts] = useState<Host[]>([]);
  const { data, isLoading, error, refetch } = useMetrics(timeRange, selectedHostId, true, customStart, customEnd);

  useEffect(() => {
    fetchHosts();
  }, []);

  const fetchHosts = async () => {
    try {
      const hostData = await hostService.getHosts();
      setHosts(hostData);
      if (hostData.length > 0 && !selectedHostId) {
        setSelectedHostId(hostData[0].id);
      }
    } catch (err) {
      console.error('Failed to fetch hosts:', err);
    }
  };
  
  const handleTimeRangeChange = (range: TimeRange, start?: string, end?: string) => {
    setTimeRange(range);
    if (range === 'custom' && start && end) {
      setCustomStart(start);
      setCustomEnd(end);
    } else {
      setCustomStart(undefined);
      setCustomEnd(undefined);
    }
  };

  const exportToCSV = (metricType: string, metricData: any[]) => {
    if (metricData.length === 0) {
      alert('No data to export');
      return;
    }

    // Create CSV content
    const headers = ['Timestamp', 'Value'];
    const rows = metricData.map(item => [
      item.timestamp,
      item.value
    ]);

    const csvContent = [
      headers.join(','),
      ...rows.map(row => row.join(','))
    ].join('\n');

    // Download CSV
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${metricType}-metrics-${timeRange}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  };

  return (
    <MainLayout>
      <div className="p-6 lg:p-8">
        {/* Page Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">Metrics</h1>
              <p className="text-gray-600">Real-time infrastructure performance metrics</p>
            </div>
            <Button
              variant="outline"
              onClick={() => refetch()}
              disabled={isLoading}
            >
              <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>

          {/* Host and Time Range Selectors */}
          <div className="flex flex-col sm:flex-row gap-4">
            {/* Host Selector */}
            <div className="flex items-center gap-3">
              <Server className="w-5 h-5 text-gray-600" />
              <select
                value={selectedHostId || ''}
                onChange={(e) => setSelectedHostId(e.target.value ? parseInt(e.target.value) : undefined)}
                className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All Hosts</option>
                {hosts.map((host) => (
                  <option key={host.id} value={host.id}>
                    {host.hostname} ({host.ip})
                  </option>
                ))}
              </select>
            </div>
            
            {/* Time Range Selector */}
            <TimeRangeSelector selected={timeRange} onChange={handleTimeRangeChange} />
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* Charts Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* CPU Usage */}
          <MetricsChart
            title="CPU Usage"
            data={data.cpu}
            dataKey="value"
            color="#3b82f6"
            type="line"
            unit="%"
            isLoading={isLoading}
            onExport={() => exportToCSV('cpu', data.cpu)}
          />

          {/* Memory Usage */}
          <MetricsChart
            title="Memory Usage"
            data={data.memory}
            dataKey="value"
            color="#10b981"
            type="area"
            unit="%"
            isLoading={isLoading}
            onExport={() => exportToCSV('memory', data.memory)}
          />

          {/* Disk Usage */}
          <MetricsChart
            title="Disk Usage"
            data={data.disk}
            dataKey="value"
            color="#f59e0b"
            type="bar"
            unit="%"
            isLoading={isLoading}
            onExport={() => exportToCSV('disk', data.disk)}
          />

          {/* Network Traffic */}
          <MetricsChart
            title="Network Traffic"
            data={data.network}
            dataKey="value"
            color="#8b5cf6"
            type="line"
            unit=" MB/s"
            isLoading={isLoading}
            onExport={() => exportToCSV('network', data.network)}
          />
        </div>

        {/* Export All Button */}
        {!isLoading && (data.cpu.length > 0 || data.memory.length > 0) && (
          <div className="mt-8 flex justify-center">
            <Button
              variant="primary"
              onClick={() => {
                exportToCSV('cpu', data.cpu);
                exportToCSV('memory', data.memory);
                exportToCSV('disk', data.disk);
                exportToCSV('network', data.network);
              }}
            >
              <Download className="w-4 h-4" />
              Export All Metrics
            </Button>
          </div>
        )}

        {/* Empty State - No Data */}
        {!isLoading && !error && data.cpu.length === 0 && data.memory.length === 0 && (
          <div className="mt-8 text-center py-12 bg-gray-50 rounded-lg border-2 border-dashed border-gray-300">
            <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <RefreshCw className="w-8 h-8 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold text-gray-900 mb-2">No Metrics Data Yet</h3>
            <p className="text-gray-600 mb-4 max-w-md mx-auto">
              Metrics will appear here once your hosts start reporting performance data.
            </p>
            <Button variant="primary" onClick={() => window.location.href = '/hosts'}>
              Add a Host to Get Started
            </Button>
          </div>
        )}
      </div>
    </MainLayout>
  );
};

export default Metrics;