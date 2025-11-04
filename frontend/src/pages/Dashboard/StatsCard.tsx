import React from 'react';
import { LucideIcon } from 'lucide-react';

interface StatsCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  color?: 'primary' | 'success' | 'warning' | 'error';
  isLoading?: boolean;
}

const StatsCard: React.FC<StatsCardProps> = ({
  title,
  value,
  icon: Icon,
  trend,
  color = 'primary',
  isLoading = false,
}) => {
  const colorClasses = {
    primary: 'bg-blue-50 text-blue-600 border-blue-200',
    success: 'bg-green-50 text-green-600 border-green-200',
    warning: 'bg-orange-50 text-orange-600 border-orange-200',
    error: 'bg-red-50 text-red-600 border-red-200',
  };

  const iconBgClasses = {
    primary: 'bg-gradient-to-br from-blue-500 to-blue-600',
    success: 'bg-gradient-to-br from-green-500 to-green-600',
    warning: 'bg-gradient-to-br from-orange-500 to-orange-600',
    error: 'bg-gradient-to-br from-red-500 to-red-600',
  };

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-3 border-l-4 border-gray-300 animate-pulse">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <div className="h-3 bg-gray-200 rounded w-1/2 mb-2"></div>
            <div className="h-6 bg-gray-200 rounded w-3/4"></div>
          </div>
          <div className="w-8 h-8 bg-gray-200 rounded-lg"></div>
        </div>
      </div>
    );
  }

  return (
    <div
      className={`bg-white rounded-lg shadow-md p-3 border-l-4 ${colorClasses[color]} hover:shadow-lg transition-all duration-300 hover:scale-[1.02] cursor-pointer group`}
    >
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-xs font-medium text-gray-600 mb-1">{title}</p>
          <h3 className="text-xl font-bold text-gray-900 mb-1">{value}</h3>
          
          {trend && (
            <div className="flex items-center gap-1">
              <span
                className={`text-xs font-medium ${
                  trend.isPositive ? 'text-green-600' : 'text-red-600'
                }`}
              >
                {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
              </span>
              <span className="text-xs text-gray-500">vs last</span>
            </div>
          )}
        </div>

        <div
          className={`w-8 h-8 ${iconBgClasses[color]} rounded-lg flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform duration-300`}
        >
          <Icon className="w-4 h-4 text-white" />
        </div>
      </div>
    </div>
  );
};

export default StatsCard;
