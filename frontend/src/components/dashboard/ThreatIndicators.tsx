import React from 'react';
import { AlertTriangle, Shield, Eye, Zap } from 'lucide-react';
import Card from '../common/Card';
import Badge from '../common/Badge';

interface ThreatIndicator {
  id: string;
  type: 'security' | 'performance' | 'availability' | 'compliance';
  severity: 'low' | 'medium' | 'high' | 'critical';
  title: string;
  description: string;
  count: number;
  timestamp: string;
}

interface ThreatIndicatorsProps {
  indicators?: ThreatIndicator[];
  isLoading: boolean;
}

const ThreatIndicators: React.FC<ThreatIndicatorsProps> = ({ indicators = [], isLoading }) => {
  const getIcon = (type: string) => {
    switch (type) {
      case 'security': return Shield;
      case 'performance': return Zap;
      case 'availability': return Eye;
      default: return AlertTriangle;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'text-red-600 bg-red-100';
      case 'high': return 'text-orange-600 bg-orange-100';
      case 'medium': return 'text-yellow-600 bg-yellow-100';
      default: return 'text-blue-600 bg-blue-100';
    }
  };

  const displayIndicators = indicators || [];

  if (isLoading) {
    return (
      <Card>
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/2 mb-4"></div>
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-center space-x-3">
                <div className="w-10 h-10 bg-gray-200 rounded-full"></div>
                <div className="flex-1">
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                  <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Card>
    );
  }

  return (
    <Card>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-900">Threat Indicators</h2>
        <Badge variant="info" size="sm">Live</Badge>
      </div>

      <div className="space-y-4">
        {displayIndicators.map((indicator) => {
          const Icon = getIcon(indicator.type);
          return (
            <div key={indicator.id} className="flex items-start space-x-4 p-3 rounded-lg hover:bg-gray-50 transition-colors">
              <div className={`p-2 rounded-full ${getSeverityColor(indicator.severity)}`}>
                <Icon className="w-5 h-5" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between mb-1">
                  <h3 className="text-sm font-semibold text-gray-900 truncate">
                    {indicator.title}
                  </h3>
                  <Badge 
                    variant={
                      indicator.severity === 'critical' ? 'error' :
                      indicator.severity === 'high' ? 'warning' :
                      indicator.severity === 'medium' ? 'info' : 'success'
                    }
                    size="sm"
                  >
                    {indicator.severity}
                  </Badge>
                </div>
                <p className="text-sm text-gray-600 mb-2">{indicator.description}</p>
                <div className="flex items-center justify-between text-xs text-gray-500">
                  <span>Count: {indicator.count}</span>
                  <span>{indicator.timestamp}</span>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {displayIndicators.length === 0 && (
        <div className="text-center py-8">
          <Shield className="w-12 h-12 text-green-500 mx-auto mb-3" />
          <p className="text-gray-600">No threats detected</p>
          <p className="text-sm text-gray-400 mt-1">Your infrastructure is secure</p>
        </div>
      )}
    </Card>
  );
};

export default ThreatIndicators;