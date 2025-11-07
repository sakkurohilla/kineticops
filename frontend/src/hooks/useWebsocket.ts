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
