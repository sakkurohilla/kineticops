import { useEffect } from 'react';
import manager from '../services/ws/manager';

type MessageHandler = (data: any) => void;

// Simple websocket hook that subscribes to the shared manager and forwards messages
export const useWebsocket = (onMessage: MessageHandler) => {
  useEffect(() => {
    const unsub = manager.subscribe((data: any) => {
      try {
        onMessage(data);
      } catch (e) {
        // swallow
      }
    });
    return () => unsub();
  }, [onMessage]);
};

export default useWebsocket;
