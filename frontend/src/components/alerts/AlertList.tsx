import React, { useState } from 'react';
import Card from '../common/Card';
import Modal from '../common/Modal';
import { Alert } from '../../types';
import AlertCard from './AlertCard';

interface AlertListProps {
  alerts: Alert[];
  onSilence?: (id: number) => Promise<any> | void;
}

const formatTime = (s?: string) => (s ? new Date(s).toLocaleString() : '');

const AlertList: React.FC<AlertListProps> = ({ alerts, onSilence }) => {
  const [selected, setSelected] = useState<Alert | null>(null);

  return (
    <div>
      <h2 className="text-lg font-semibold mb-4">Alerts</h2>

      {alerts.length === 0 ? (
        <Card className="p-6 text-center">No alerts yet</Card>
      ) : (
        <ul className="space-y-3">
          {alerts.map((a) => (
            <li key={a.id}>
              <div onClick={() => setSelected(a)} className="cursor-pointer">
                <Card className="p-4 hover:bg-gray-50">
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className={`px-2 py-1 text-xs font-medium rounded ${a.severity === 'critical' ? 'bg-red-100 text-red-800' : a.severity === 'warning' ? 'bg-yellow-100 text-yellow-800' : 'bg-blue-100 text-blue-800'}`}>
                          {a.severity.toUpperCase()}
                        </span>
                        <h3 className="text-sm font-medium">{a.message}</h3>
                      </div>
                      <p className="text-sm text-gray-500 mt-1">Host: {a.host_name || a.host_id} • {formatTime(a.triggered_at || a.created_at)}</p>
                    </div>
                    <div className="text-sm text-gray-400">{a.status}</div>
                  </div>
                </Card>
              </div>
            </li>
          ))}
        </ul>
      )}

      {selected && (
        <Modal isOpen={!!selected} onClose={() => setSelected(null)} title={`Alert #${selected.id}`} size="md">
          <AlertCard alert={selected} onSilence={async (id: number) => {
            if (onSilence) await onSilence(id);
            setSelected(null);
          }} />
        </Modal>
      )}
    </div>
  );
};

export default AlertList;
