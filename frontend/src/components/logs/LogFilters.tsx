import React, { useEffect, useState } from 'react';
import Input from '../common/Input';
import logsService from '../../services/api/logsService';
import hostService from '../../services/api/hostService';

interface LogFiltersProps {
  filters: {
    level?: string;
    source?: string;
    host_id?: number;
    service?: string;
    startDate?: string;
    endDate?: string;
    search?: string;
  };
  onFiltersChange: (filters: any) => void;
}

const LogFilters: React.FC<LogFiltersProps> = ({ filters, onFiltersChange }) => {
  const handleChange = (key: string, value: any) => {
    onFiltersChange({ ...filters, [key]: value });
  };

  const [sources, setSources] = useState<string[]>([]);
  const [levels, setLevels] = useState<string[]>([]);
  const [hosts, setHosts] = useState<any[]>([]);
  const [servicesList, setServicesList] = useState<string[]>([]);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const res = await logsService.getSources();
        if (!mounted) return;
  setSources(res.sources || []);
  setLevels(res.levels || []);
        // also fetch hosts for the host selector
        try {
          const h = await hostService.getAllHosts();
          if (mounted) setHosts(h || []);
        } catch (e) {
          // ignore
        }
      } catch (err) {
        // ignore - optional UX enhancement
      }
    })();
    return () => {
      mounted = false;
    };
  }, []);

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
          {(levels && levels.length ? levels : ['debug', 'info', 'warn', 'error']).map((lvl) => (
            <option key={lvl} value={lvl}>
              {lvl.charAt(0).toUpperCase() + lvl.slice(1)}
            </option>
          ))}
        </select>
      </div>

      {/* Source Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Source
        </label>
        <select
          value={filters.source || ''}
          onChange={(e) => handleChange('source', e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">All Sources</option>
          {sources.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
      </div>

      {/* Host Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Host</label>
        <select
          value={filters.host_id || ''}
          onChange={async (e) => {
            const v = e.target.value;
            handleChange('host_id', v ? Number(v) : undefined);
            // fetch services for this host
            if (v) {
              try {
                const s = await hostService.getHostServices(Number(v));
                setServicesList((s || []).map((x: any) => x.service_name || x.ServiceName || x.serviceName || x.ServiceName));
              } catch (err) {
                setServicesList([]);
              }
            } else {
              setServicesList([]);
            }
          }}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">All Hosts</option>
          {hosts.map((h) => (
            <option key={h.id} value={h.id}>
              {h.hostname}
            </option>
          ))}
        </select>
      </div>

      {/* Service Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Service</label>
        <select
          value={(filters as any).service || ''}
          onChange={(e) => handleChange('service', e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">All Services</option>
          {servicesList.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
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
