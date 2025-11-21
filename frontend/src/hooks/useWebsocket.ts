import { useEffect, useRef } from 'react';
import manager from '../services/ws/manager';

type MessageHandler = (data: any) => void;

// Improved websocket hook: subscribe only once and keep a ref to the latest handler.
// This avoids re-subscribing on every render when components pass inline/unstable
// callbacks, which previously caused subscribe/unsubscribe churn and frequent
// websocket reconnects.
export const useWebsocket = (onMessage: MessageHandler) => {
  const handlerRef = useRef<MessageHandler>(onMessage);

  // keep the ref up-to-date with the latest handler so the single subscription
  // can always call the current callback without re-subscribing.
  useEffect(() => {
    handlerRef.current = onMessage;
  }, [onMessage]);

  useEffect(() => {
    const unsub = manager.subscribe((data: any) => {
      try {
        // Mark clearly empty metric frames as placeholders to skip average recompute.
        if (data && data.type === 'metric') {
          const isPlaceholder = (
            (data.cpu_usage === 0 || data.cpu_usage === null || data.cpu_usage === undefined) &&
            (data.memory_usage === null || data.memory_usage === undefined) &&
            (data.disk_usage === null || data.disk_usage === undefined) &&
            !data.memory_total && !data.memory_total_bytes &&
            !data.disk_total && !data.disk_total_bytes
          );
          if (isPlaceholder) {
            data._placeholder = true;
          }
        }
        handlerRef.current(data);
      } catch (e) {
        // swallow
      }
    });
    return () => unsub();
    // subscribe only once on mount
  }, []);
};

export default useWebsocket;
