import React from 'react';
import Input from '../common/Input';

interface LogFiltersProps {
  filters: {
    level?: string;
    source?: string;
    startDate?: string;
    endDate?: string;
    search?: string;
  };
  onFiltersChange: (filters: any) => void;
}

const LogFilters: React.FC<LogFiltersProps> = ({ filters, onFiltersChange }) => {
  const handleChange = (key: string, value: string) => {
    onFiltersChange({ ...filters, [key]: value });
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* Log Level Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Log Level
        </label>
        <select
          value={filters.level || ''}
          onChange={(e) => handleChange('level', e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">All Levels</option>
          <option value="debug">Debug</option>
          <option value="info">Info</option>
          <option value="warn">Warning</option>
          <option value="error">Error</option>
        </select>
      </div>

      {/* Source Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Source
        </label>
        <Input
          type="text"
          placeholder="Filter by source..."
          value={filters.source || ''}
          onChange={(e) => handleChange('source', e.target.value)}
        />
      </div>

      {/* Start Date Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Start Date
        </label>
        <Input
          type="datetime-local"
          value={filters.startDate || ''}
          onChange={(e) => handleChange('startDate', e.target.value)}
        />
      </div>

      {/* End Date Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          End Date
        </label>
        <Input
          type="datetime-local"
          value={filters.endDate || ''}
          onChange={(e) => handleChange('endDate', e.target.value)}
        />
      </div>
    </div>
  );
};

export default LogFilters;
