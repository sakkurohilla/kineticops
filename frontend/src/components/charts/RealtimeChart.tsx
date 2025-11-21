import React, { useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import useWebsocket from '../../hooks/useWebsocket';

interface MetricDataPoint {
  timestamp: string;
  value: number;
}

interface RealtimeChartProps {
  hostId?: number;
  metricName: string;
  title: string;
  color?: string;
  maxDataPoints?: number;
  unit?: string;
}

const RealtimeChart: React.FC<RealtimeChartProps> = ({
  hostId,
  metricName,
  title,
  color = '#3b82f6',
  maxDataPoints = 20,
  unit = '%',
}) => {
  const [data, setData] = useState<MetricDataPoint[]>([]);

  // Listen to WebSocket updates
  useWebsocket((payload: any) => {
    if (payload.type === 'metric_update' || payload.type === 'metric') {
      const metric = payload.data || payload;
      
      // Filter for the specific metric we're tracking
      if (metric.name === metricName && (!hostId || metric.host_id === hostId)) {
        const newPoint: MetricDataPoint = {
          timestamp: new Date(metric.timestamp).toLocaleTimeString(),
          value: parseFloat(metric.value) || 0,
        };

        setData((prev: MetricDataPoint[]) => {
          const updated = [...prev, newPoint];
          // Keep only the last N data points
          if (updated.length > maxDataPoints) {
            return updated.slice(updated.length - maxDataPoints);
          }
          return updated;
        });
      }
    }
  });

  return (
    <div className="bg-white rounded-lg shadow-lg p-4">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">{title}</h3>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
          <XAxis
            dataKey="timestamp"
            stroke="#6b7280"
            tick={{ fontSize: 12 }}
            tickFormatter={(value) => value.split(':').slice(0, 2).join(':')}
          />
          <YAxis
            stroke="#6b7280"
            tick={{ fontSize: 12 }}
            domain={[0, 100]}
            tickFormatter={(value) => `${value}${unit}`}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: '#fff',
              border: '1px solid #e5e7eb',
              borderRadius: '0.5rem',
            }}
            formatter={(value: number) => [`${value.toFixed(2)}${unit}`, metricName]}
          />
          <Legend />
          <Line
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={2}
            dot={false}
            name={metricName}
            animationDuration={300}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};

export default RealtimeChart;
