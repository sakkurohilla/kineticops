import React from 'react';
import Card from '../common/Card';
import { Alert } from '../../types';

interface Props {
  alert: Alert;
  onSilence?: (id: number) => Promise<any> | void;
}

const AlertCard: React.FC<Props> = ({ alert, onSilence }) => {
  const handleSilence = async () => {
    if (!onSilence) return;
    await onSilence(alert.id);
  };

  return (
    <div>
      <Card className="mb-4">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-3">
              <h3 className="text-lg font-semibold">{alert.message}</h3>
              <span className={`px-2 py-1 rounded text-xs ${alert.severity === 'critical' ? 'bg-red-100 text-red-800' : alert.severity === 'warning' ? 'bg-yellow-100 text-yellow-800' : 'bg-blue-100 text-blue-800'}`}>
                {alert.severity}
              </span>
            </div>
            <p className="text-sm text-gray-500 mt-2">Rule: {alert.rule_id || '—'}</p>
            <p className="text-sm text-gray-500">Host: {alert.host_name || alert.host_id}</p>
            <p className="text-sm text-gray-500">Status: {alert.status}</p>
            <p className="text-sm text-gray-500">Triggered: {alert.triggered_at || alert.created_at}</p>
          </div>
          <div className="flex flex-col items-end gap-2">
            <button onClick={handleSilence} className="px-3 py-2 rounded bg-gray-100 hover:bg-gray-200 text-sm">
              Silence
            </button>
            <button className="px-3 py-2 rounded bg-blue-600 text-white text-sm">Escalate</button>
          </div>
        </div>
      </Card>

      <Card>
        <h4 className="font-medium mb-2">Details</h4>
        <pre className="text-sm text-gray-700 whitespace-pre-wrap">{JSON.stringify(alert, null, 2)}</pre>
      </Card>
    </div>
  );
};

export default AlertCard;
