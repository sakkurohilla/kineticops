import React, { useState, useEffect } from 'react';
import { Wifi, WifiOff, Clock, Shield } from 'lucide-react';
import Card from '../common/Card';
import Badge from '../common/Badge';

interface RealTimeStatusProps {
  isConnected?: boolean;
}

const RealTimeStatus: React.FC<RealTimeStatusProps> = ({ isConnected = true }) => {
  const [lastUpdate, setLastUpdate] = useState(new Date());
  const [uptime] = useState('--');

  useEffect(() => {
    const interval = setInterval(() => {
      setLastUpdate(new Date());
    }, 30000); // Update every 30 seconds

    return () => clearInterval(interval);
  }, []);

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('en-US', { 
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  return (
    <Card className="bg-gradient-to-r from-slate-50 to-gray-50">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-bold text-gray-900">System Status</h3>
        <div className="flex items-center space-x-2">
          {isConnected ? (
            <>
              <Wifi className="w-5 h-5 text-green-500" />
              <Badge variant="success" size="sm">Online</Badge>
            </>
          ) : (
            <>
              <WifiOff className="w-5 h-5 text-red-500" />
              <Badge variant="error" size="sm">Offline</Badge>
            </>
          )}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="text-center p-3 bg-white rounded-lg border">
          <div className="flex items-center justify-center mb-2">
            <Clock className="w-4 h-4 text-blue-500 mr-2" />
            <span className="text-sm font-medium text-gray-700">Last Update</span>
          </div>
          <div className="text-lg font-bold text-gray-900">{formatTime(lastUpdate)}</div>
        </div>

        <div className="text-center p-3 bg-white rounded-lg border">
          <div className="flex items-center justify-center mb-2">
            <Shield className="w-4 h-4 text-green-500 mr-2" />
            <span className="text-sm font-medium text-gray-700">Uptime</span>
          </div>
          <div className="text-lg font-bold text-green-600">{uptime}</div>
        </div>
      </div>

      <div className="mt-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-blue-900">Monitoring Engine</span>
          <Badge variant={isConnected ? "success" : "error"} size="sm">{isConnected ? "Active" : "Inactive"}</Badge>
        </div>
        <p className="text-xs text-blue-700 mt-1">Real-time monitoring {isConnected ? 'enabled' : 'disabled'}</p>
      </div>
    </Card>
  );
};

export default RealTimeStatus;