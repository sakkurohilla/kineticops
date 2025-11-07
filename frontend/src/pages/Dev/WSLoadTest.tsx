import React, { useState } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import apiClient from '../../services/api/client';

const WSLoadTest: React.FC = () => {
  const [hostId, setHostId] = useState<number>(1);
  const [count, setCount] = useState<number>(1000);
  const [intervalMs, setIntervalMs] = useState<number>(1);
  const [status, setStatus] = useState<string>('idle');

  const sendBurst = async () => {
    setStatus('sending');
    try {
      const res: any = await apiClient.post('/internal/debug/ws/burst', {
        host_id: hostId,
        count,
        interval_ms: intervalMs,
      });
      setStatus(res.msg || 'sent');
    } catch (e: any) {
      setStatus(e?.message || 'error');
    }
  };

  return (
    <MainLayout>
      <div className="p-6">
        <h2 className="text-xl font-bold mb-4">WebSocket Load Test (Dev)</h2>
        <div className="space-y-3 max-w-md">
          <label className="block">
            Host ID
            <input type="number" value={hostId} onChange={e => setHostId(Number(e.target.value))} className="w-full p-2 border rounded mt-1" />
          </label>
          <label className="block">
            Message count
            <input type="number" value={count} onChange={e => setCount(Number(e.target.value))} className="w-full p-2 border rounded mt-1" />
          </label>
          <label className="block">
            Interval (ms)
            <input type="number" value={intervalMs} onChange={e => setIntervalMs(Number(e.target.value))} className="w-full p-2 border rounded mt-1" />
          </label>
          <div>
            <button onClick={sendBurst} className="px-4 py-2 bg-blue-600 text-white rounded">Send Burst</button>
            <span className="ml-4">Status: {status}</span>
          </div>
        </div>
      </div>
    </MainLayout>
  );
};

export default WSLoadTest;
