import React, { useEffect, useState } from 'react';
import wsStatus from '../../utils/wsStatus';
import manager from '../../services/ws/manager';

const WsOverlay: React.FC = () => {
  const [status, setStatus] = useState<string>('disconnected');
  const [info, setInfo] = useState<string | undefined>(undefined);
  const [subscribers, setSubscribers] = useState<number>(0);
  const [events, setEvents] = useState<string[]>([]);

  useEffect(() => {
    const unsub = wsStatus.subscribeWsStatus((s, i) => {
      setStatus(s);
      setInfo(i);
      setEvents((prev) => [`${new Date().toLocaleTimeString()} ${s}${i ? ' - ' + i : ''}`, ...prev].slice(0, 10));
    });

    const t = setInterval(() => {
      try {
        setSubscribers(manager.getSubscriberCount());
      } catch (e) {
        // ignore
      }
    }, 1000);

    return () => {
      unsub();
      clearInterval(t);
    };
  }, []);

  return (
    <div className="fixed right-4 bottom-24 z-50 w-72 pointer-events-auto">
      <div className="bg-white/90 backdrop-blur-md border border-gray-200 rounded-lg p-3 shadow-lg text-xs">
        <div className="flex items-center justify-between mb-2">
          <div className="font-semibold">WebSocket</div>
          <div className={`px-2 py-0.5 rounded text-white text-[10px] ${status === 'connected' ? 'bg-green-600' : status === 'connecting' || status === 'reconnecting' ? 'bg-yellow-600' : 'bg-gray-500'}`}>
            {status}
          </div>
        </div>
        <div className="text-[11px] text-gray-700 mb-2">
          Subscribers: <span className="font-medium">{subscribers}</span>
        </div>
        {info && (
          <div className="text-[11px] text-gray-600 mb-2">Info: {info}</div>
        )}
        <div className="h-24 overflow-auto">
          {events.map((e, i) => (
            <div key={i} className="text-[11px] text-gray-600">{e}</div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default WsOverlay;
