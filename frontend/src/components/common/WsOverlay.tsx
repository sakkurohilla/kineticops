import React, { useEffect, useState } from 'react';
import wsStatus from '../../utils/wsStatus';
import manager from '../../services/ws/manager';

const WsOverlay: React.FC = () => {
  const [status, setStatus] = useState<string>('disconnected');
  const [info, setInfo] = useState<string | undefined>(undefined);
  const [subscribers, setSubscribers] = useState<number>(0);
  const [events, setEvents] = useState<string[]>([]);
  const [isMinimized, setIsMinimized] = useState(() => {
    // Load minimize state from localStorage
    const saved = localStorage.getItem('wsOverlayMinimized');
    return saved ? JSON.parse(saved) : false;
  });

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

  // Save minimize state to localStorage
  const toggleMinimize = () => {
    const newState = !isMinimized;
    setIsMinimized(newState);
    localStorage.setItem('wsOverlayMinimized', JSON.stringify(newState));
  };

  const statusColor = status === 'connected' ? 'bg-green-600' : status === 'connecting' || status === 'reconnecting' ? 'bg-yellow-600' : 'bg-gray-500';

  return (
    <div className="fixed right-4 bottom-24 z-50 pointer-events-auto">
      {isMinimized ? (
        // Minimized view - compact button
        <button
          onClick={toggleMinimize}
          className="bg-white/90 backdrop-blur-md border border-gray-200 rounded-lg px-3 py-2 shadow-lg hover:bg-white transition-colors flex items-center gap-2"
          title="Expand WebSocket debug panel"
        >
          <div className={`w-2 h-2 rounded-full ${statusColor} ${status === 'connected' ? 'animate-pulse' : ''}`}></div>
          <span className="text-xs font-medium text-gray-700">WS</span>
          <span className="text-[10px] text-gray-500">({events.length})</span>
        </button>
      ) : (
        // Expanded view - full panel
        <div className="w-72 bg-white/90 backdrop-blur-md border border-gray-200 rounded-lg shadow-lg text-xs">
          {/* Header with minimize button */}
          <div className="flex items-center justify-between p-3 border-b border-gray-100">
            <div className="font-semibold text-gray-700">WebSocket Debug</div>
            <div className="flex items-center gap-2">
              <div className={`px-2 py-0.5 rounded text-white text-[10px] ${statusColor}`}>
                {status}
              </div>
              <button
                onClick={toggleMinimize}
                className="text-gray-400 hover:text-gray-600 p-1 rounded hover:bg-gray-100 transition-colors"
                title="Minimize"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
            </div>
          </div>

          {/* Content */}
          <div className="p-3">
            <div className="text-[11px] text-gray-700 mb-2">
              Subscribers: <span className="font-medium">{subscribers}</span>
            </div>
            {info && (
              <div className="text-[11px] text-gray-600 mb-2">Info: {info}</div>
            )}
            
            {/* Events log */}
            <div className="mt-2">
              <div className="text-[10px] text-gray-500 mb-1 flex items-center justify-between">
                <span>Recent Events</span>
                {events.length > 0 && (
                  <button
                    onClick={() => setEvents([])}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    Clear
                  </button>
                )}
              </div>
              <div className="h-24 overflow-auto bg-gray-50 rounded p-2">
                {events.length === 0 ? (
                  <div className="text-[11px] text-gray-400 text-center py-4">No events yet</div>
                ) : (
                  events.map((e, i) => (
                    <div key={i} className="text-[11px] text-gray-600 mb-1">{e}</div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default WsOverlay;
