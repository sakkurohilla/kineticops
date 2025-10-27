import React from 'react';
import { Log } from '../../types';
import Badge from '../common/Badge';
import { format } from 'date-fns';
import { Eye } from 'lucide-react';

interface LogsTableProps {
  logs: Log[];
  onLogClick: (log: Log) => void;
  isLoading?: boolean;
}

const LogsTable: React.FC<LogsTableProps> = ({ logs, onLogClick, isLoading }) => {
  const getLevelBadgeVariant = (level: string): 'success' | 'warning' | 'error' | 'info' => {
    switch (level.toLowerCase()) {
      case 'error':
        return 'error';
      case 'warn':
      case 'warning':
        return 'warning';
      case 'info':
        return 'info';
      default:
        return 'success';
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96 bg-white rounded-lg shadow">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading logs...</p>
        </div>
      </div>
    );
  }

  if (logs.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow">
        <div className="text-center py-12">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No logs found</h3>
          <p className="mt-1 text-sm text-gray-500">Try adjusting your filters or date range.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      {/* Header */}
      <div className="bg-gray-50 px-6 py-3 border-b border-gray-200">
        <div className="grid grid-cols-12 gap-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
          <div className="col-span-2">Timestamp</div>
          <div className="col-span-1">Level</div>
          <div className="col-span-2">Host</div>
          <div className="col-span-2">Source</div>
          <div className="col-span-4">Message</div>
          <div className="col-span-1">Action</div>
        </div>
      </div>

      {/* Table Body - Scrollable */}
      <div className="overflow-y-auto" style={{ maxHeight: '600px' }}>
        {logs.map((log, index) => (
          <div
            key={log.id || index}
            className="flex items-center border-b border-gray-200 hover:bg-gray-50 cursor-pointer px-6 py-4 transition-colors"
            onClick={() => onLogClick(log)}
          >
            <div className="grid grid-cols-12 gap-4 items-center w-full">
              {/* Timestamp */}
              <div className="col-span-2 text-sm text-gray-900 font-mono">
                {format(new Date(log.timestamp), 'MMM dd, HH:mm:ss')}
              </div>

              {/* Level */}
              <div className="col-span-1">
                <Badge variant={getLevelBadgeVariant(log.level)}>
                  {log.level.toUpperCase()}
                </Badge>
              </div>

              {/* Host */}
              <div className="col-span-2 text-sm text-gray-900 truncate" title={log.host_id}>
                {log.host_id}
              </div>

              {/* Source */}
              <div className="col-span-2 text-sm text-gray-900 font-mono truncate" title={log.source}>
                {log.source}
              </div>

              {/* Message */}
              <div className="col-span-4 text-sm text-gray-600 truncate" title={log.message}>
                {log.message}
              </div>

              {/* Action */}
              <div className="col-span-1 flex justify-center">
                <Eye className="h-5 w-5 text-gray-400 hover:text-blue-600" />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Footer with count */}
      <div className="bg-gray-50 px-6 py-3 border-t border-gray-200">
        <p className="text-sm text-gray-700">
          Showing <span className="font-medium">{logs.length}</span> log entries
        </p>
      </div>
    </div>
  );
};

export default LogsTable;
