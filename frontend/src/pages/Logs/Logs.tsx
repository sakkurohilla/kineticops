import React, { useState } from 'react';
import { Search, Filter, RefreshCw, Download, Play, Pause } from 'lucide-react';
import MainLayout from '../../components/layout/MainLayout';
import Input from '../../components/common/Input';
import Button from '../../components/common/Button';
import LogsTable from '../../components/logs/LogsTable';
import LogFilters from '../../components/logs/LogFilters';
import LogDetails from '../../components/logs/LogDetails';
import { useLogs } from '../../hooks/useLogs';
import { Log } from '../../types';

const Logs: React.FC = () => {
  const [searchQuery, setSearchQuery] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [selectedLog, setSelectedLog] = useState<Log | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const {
    logs,
    isLoading,
    filters,
    setFilters,
    isTailing,
    toggleTailing,
    refetch,
  } = useLogs();

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
    setFilters({ ...filters, search: e.target.value });
  };

  const handleLogClick = (log: Log) => {
    setSelectedLog(log);
    setIsModalOpen(true);
  };

  const handleExport = () => {
    if (logs.length === 0) {
      alert('No data to export');
      return;
    }

    const headers = ['Timestamp', 'Level', 'Host', 'Source', 'Message'];
    const csvContent = [
      headers.join(','),
      ...logs.map(log =>
        [
          log.timestamp,
          log.level,
          log.host_id,
          log.source,
          `"${log.message.replace(/"/g, '""')}"`,
        ].join(',')
      ),
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `logs-${new Date().toISOString()}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
  };

  return (
    <MainLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Logs</h1>
          <div className="flex items-center gap-3">
            <Button
              variant="outline"
              onClick={refetch}
              className="flex items-center gap-2"
            >
              <RefreshCw className="h-4 w-4" />
              Refresh
            </Button>
            <Button
              variant="outline"
              onClick={handleExport}
              className="flex items-center gap-2"
            >
              <Download className="h-4 w-4" />
              Export
            </Button>
            <Button
              variant={isTailing ? 'error' : 'primary'}
              onClick={toggleTailing}
              className="flex items-center gap-2"
            >
              {isTailing ? (
                <>
                  <Pause className="h-4 w-4" />
                  Stop Tailing
                </>
              ) : (
                <>
                  <Play className="h-4 w-4" />
                  Start Tailing
                </>
              )}
            </Button>
          </div>
        </div>

        {/* Search and Filters */}
        <div className="bg-white rounded-lg shadow p-4">
          <div className="flex items-center gap-4">
            <div className="flex-1">
              <Input
                type="text"
                placeholder="Search logs..."
                value={searchQuery}
                onChange={handleSearch}
                icon={<Search className="h-5 w-5 text-gray-400" />}
              />
            </div>
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              className="flex items-center gap-2"
            >
              <Filter className="h-4 w-4" />
              Filters
            </Button>
          </div>

          {showFilters && (
            <div className="mt-4 pt-4 border-t">
              <LogFilters filters={filters} onFiltersChange={setFilters} />
            </div>
          )}
        </div>

        {/* Logs Table */}
        <LogsTable
          logs={logs}
          onLogClick={handleLogClick}
          isLoading={isLoading}
        />

        {/* Log Details Modal */}
        {isModalOpen && selectedLog && (
          <LogDetails
            log={selectedLog}
            onClose={() => {
              setIsModalOpen(false);
              setSelectedLog(null);
            }}
          />
        )}
      </div>
    </MainLayout>
  );
};

export default Logs;
