import React from 'react';
import { Shield } from 'lucide-react';
import Card from '../common/Card';

interface SystemOverviewProps {
  stats: {
    totalHosts: number;
    onlineHosts: number;
    warnings: number;
    critical: number;
  };
  isLoading: boolean;
}

const SystemOverview: React.FC<SystemOverviewProps> = ({ stats, isLoading }) => {
  const healthPercentage = stats.totalHosts > 0 
    ? Math.round((stats.onlineHosts / stats.totalHosts) * 100) 
    : 0;

  const getHealthColor = (percentage: number) => {
    if (percentage >= 90) return 'text-green-600';
    if (percentage >= 70) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getHealthBgColor = (percentage: number) => {
    if (percentage >= 90) return 'bg-green-100';
    if (percentage >= 70) return 'bg-yellow-100';
    return 'bg-red-100';
  };

  if (isLoading) {
    return (
      <Card>
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
          <div className="space-y-3">
            <div className="h-4 bg-gray-200 rounded"></div>
            <div className="h-4 bg-gray-200 rounded w-5/6"></div>
            <div className="h-4 bg-gray-200 rounded w-4/6"></div>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <Card className="bg-gradient-to-br from-blue-50 to-indigo-100 border-blue-200">
      <div className="flex items-start justify-between mb-6">
        <div>
          <h2 className="text-xl font-bold text-gray-900 mb-2">System Health Overview</h2>
          <p className="text-sm text-gray-600">Real-time infrastructure status</p>
        </div>
        <div className={`p-3 rounded-full ${getHealthBgColor(healthPercentage)}`}>
          <Shield className={`w-6 h-6 ${getHealthColor(healthPercentage)}`} />
        </div>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="text-center">
          <div className="text-2xl font-bold text-gray-900">{stats.totalHosts}</div>
          <div className="text-sm text-gray-600">Total Hosts</div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold text-green-600">{stats.onlineHosts}</div>
          <div className="text-sm text-gray-600">Online</div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold text-yellow-600">{stats.warnings}</div>
          <div className="text-sm text-gray-600">Warnings</div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold text-red-600">{stats.critical}</div>
          <div className="text-sm text-gray-600">Critical</div>
        </div>
      </div>

      <div className="bg-white rounded-lg p-4 border border-blue-200">
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm font-medium text-gray-700">Overall Health</span>
          <span className={`text-lg font-bold ${getHealthColor(healthPercentage)}`}>
            {healthPercentage}%
          </span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-3">
          <div 
            className={`h-3 rounded-full transition-all duration-500 ${
              healthPercentage >= 90 ? 'bg-green-500' :
              healthPercentage >= 70 ? 'bg-yellow-500' : 'bg-red-500'
            }`}
            style={{ width: `${healthPercentage}%` }}
          ></div>
        </div>
      </div>
    </Card>
  );
};

export default SystemOverview;