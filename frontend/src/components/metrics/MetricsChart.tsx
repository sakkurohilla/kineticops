import React from 'react';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { format } from 'date-fns';
import Card from '../common/Card';
import { Download, Activity } from 'lucide-react';

export type ChartType = 'line' | 'area' | 'bar';

interface MetricsChartProps {
  title: string;
  data: any[];
  dataKey: string;
  color: string;
  type?: ChartType;
  unit?: string;
  isLoading?: boolean;
  onExport?: () => void;
}

const MetricsChart: React.FC<MetricsChartProps> = ({
  title,
  data,
  dataKey,
  color,
  type = 'line',
  unit = '%',
  isLoading = false,
  onExport,
}) => {
  const formatXAxis = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffHours = Math.abs(now.getTime() - date.getTime()) / (1000 * 60 * 60);
      
      // Format based on time range
      if (diffHours > 168) { // > 7 days
        return format(date, 'MM/dd');
      } else if (diffHours > 24) { // > 1 day
        return format(date, 'MM/dd HH:mm');
      } else {
        return format(date, 'HH:mm');
      }
    } catch {
      return timestamp;
    }
  };

  const formatTooltip = (value: any) => {
    return `${value}${unit}`;
  };

  if (isLoading) {
    return (
      <Card>
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
          <div className="h-64 bg-gray-100 rounded"></div>
        </div>
      </Card>
    );
  }

  const renderChart = () => {
    const commonProps = {
      data,
      margin: { top: 10, right: 30, left: 0, bottom: 0 },
    };

    switch (type) {
      case 'area':
        return (
          <AreaChart {...commonProps}>
            <defs>
              <linearGradient id={`gradient-${dataKey}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor={color} stopOpacity={0.3} />
                <stop offset="95%" stopColor={color} stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
            {/* Use category X axis so points are evenly spaced; compute interval to avoid crowded ticks */}
            <XAxis dataKey="timestamp" tickFormatter={formatXAxis} stroke="#9ca3af" type="category" interval={data.length > 6 ? Math.floor(data.length / 6) : 0} />
            <YAxis stroke="#9ca3af" />
            <Tooltip 
              formatter={formatTooltip}
              labelFormatter={(label) => {
                try {
                  return format(new Date(label), 'MMM dd, yyyy HH:mm:ss');
                } catch {
                  return label;
                }
              }}
            />
            <Area
              type="monotone"
              dataKey={dataKey}
              stroke={color}
              strokeWidth={2}
              fill={`url(#gradient-${dataKey})`}
            />
          </AreaChart>
        );

      case 'bar':
        return (
          <BarChart {...commonProps}>
            <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
            {/* Use category X axis so points are evenly spaced; compute interval to avoid crowded ticks */}
            <XAxis dataKey="timestamp" tickFormatter={formatXAxis} stroke="#9ca3af" type="category" interval={data.length > 6 ? Math.floor(data.length / 6) : 0} />
            <YAxis stroke="#9ca3af" />
            <Tooltip 
              formatter={formatTooltip}
              labelFormatter={(label) => {
                try {
                  return format(new Date(label), 'MMM dd, yyyy HH:mm:ss');
                } catch {
                  return label;
                }
              }}
            />
            <Bar dataKey={dataKey} fill={color} radius={[8, 8, 0, 0]} />
          </BarChart>
        );

      default: // line
        return (
          <LineChart {...commonProps}>
            <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
            {/* Use category X axis so points are evenly spaced; compute interval to avoid crowded ticks */}
            <XAxis dataKey="timestamp" tickFormatter={formatXAxis} stroke="#9ca3af" type="category" interval={data.length > 6 ? Math.floor(data.length / 6) : 0} />
            <YAxis stroke="#9ca3af" />
            <Tooltip 
              formatter={formatTooltip}
              labelFormatter={(label) => {
                try {
                  return format(new Date(label), 'MMM dd, yyyy HH:mm:ss');
                } catch {
                  return label;
                }
              }}
            />
            <Legend />
            <Line
              type="monotone"
              dataKey={dataKey}
              stroke={color}
              strokeWidth={2}
              dot={{ fill: color, r: 4 }}
              activeDot={{ r: 6 }}
            />
          </LineChart>
        );
    }
  };

  return (
    <Card className="hover:shadow-lg transition-shadow duration-300">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-bold text-gray-900">{title}</h3>
        {onExport && (
          <button
            onClick={onExport}
            className="p-2 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
            title="Export to CSV"
          >
            <Download className="w-4 h-4" />
          </button>
        )}
      </div>

      {data.length === 0 ? (
        <div className="h-64 flex items-center justify-center text-gray-500">
          <div className="text-center">
            <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <Activity className="w-8 h-8 text-gray-400" />
            </div>
            <p className="text-lg mb-2 font-medium">No {title.toLowerCase()} data</p>
            <p className="text-sm text-gray-400">Data will appear when hosts start reporting metrics</p>
          </div>
        </div>
      ) : (
        <div style={{ width: '100%', height: 300 }}>
          <ResponsiveContainer width="100%" height={300}>
            {renderChart()}
          </ResponsiveContainer>
        </div>
      )}
    </Card>
  );
};

export default MetricsChart;
