import authService from '../auth/authService';
import config from '../../config/config';
import { BASE_URL as API_BASE } from '../api/client';
import wsStatus from '../../utils/wsStatus';

type MessageHandler = (data: any) => void;

let ws: WebSocket | null = null;
let subscribers = new Set<MessageHandler>();
let connectAttempts = 0;
let reconnectTimer: number | null = null;

function deriveBaseWs() {
  let baseWs = config.wsUrl.replace(/\/?$/, '');
  try {
    if (API_BASE) {
      const apiUrl = new URL(API_BASE);
      const proto = apiUrl.protocol === 'https:' ? 'wss' : 'ws';
      baseWs = `${proto}://${apiUrl.host}/ws`;
    } else if ((config as any).apiBaseUrl) {
      const apiUrl = new URL((config as any).apiBaseUrl);
      const proto = apiUrl.protocol === 'https:' ? 'wss' : 'ws';
      baseWs = `${proto}://${apiUrl.host}/ws`;
    }
  } catch (e) {
    console.warn('[wsManager] failed to derive ws base from API_BASE/config, using configured wsUrl', e);
  }
  return baseWs;
}

function notifyAll(data: any) {
  subscribers.forEach((s) => {
    try {
      s(data);
    } catch (e) {
      // ignore
    }
  });
}

function scheduleReconnect() {
  const base = 1000;
  const max = 30000;
  const backoff = Math.min(max, base * Math.pow(2, Math.min(6, connectAttempts)));
  const jitter = Math.floor(Math.random() * 1000);
  const wait = backoff + jitter;
  if (reconnectTimer) window.clearTimeout(reconnectTimer);
  reconnectTimer = window.setTimeout(() => {
    connect();
  }, wait);
  wsStatus.setWsStatus('reconnecting', String(wait));
}

function connect() {
  if (ws) return; // already connected/connecting
  connectAttempts += 1;
  const token = authService.getToken();
  const baseWs = deriveBaseWs();
  wsStatus.setWsStatus('connecting', baseWs);
  const separator = baseWs.includes('?') ? '&' : '?';
  const url = `${baseWs}${separator}token=${encodeURIComponent(token || '')}`;
  try {
    ws = new WebSocket(url);
  } catch (err) {
    console.error('[wsManager] WebSocket connect error', err);
    scheduleReconnect();
    return;
  }

  ws.onopen = () => {
    connectAttempts = 0;
    console.log('[wsManager] connected to', url);
    wsStatus.setWsStatus('connected');
  };

  ws.onmessage = (ev) => {
    try {
      const parsed = JSON.parse(ev.data);
      notifyAll(parsed);
    } catch (e) {
      // ignore non-json
    }
  };

  ws.onclose = (ev) => {
    ws = null;
    if (ev && ev.code !== 1000) {
      console.warn('[wsManager] connection closed, scheduling reconnect', ev.code);
      scheduleReconnect();
    } else {
      wsStatus.setWsStatus('disconnected');
    }
  };

  ws.onerror = (e) => {
    console.error('[wsManager] socket error', e);
    wsStatus.setWsStatus('error', String(e));
  };
}

export function subscribe(handler: MessageHandler) {
  subscribers.add(handler);
  // ensure connection exists
  if (!ws) connect();
  return () => {
    subscribers.delete(handler);
    // if no subscribers, close the socket after a longer idle delay (allow quick navigation without closing)
      if (subscribers.size === 0 && ws) {
        // keep socket alive for 5 minutes by default when there are no subscribers
        // this prevents quick navigation or transient unmounts from closing the socket
        // and causing reconnect churn.
        window.setTimeout(() => {
          if (subscribers.size === 0 && ws) {
            ws?.close();
            ws = null;
            wsStatus.setWsStatus('disconnected');
          }
        }, 5 * 60 * 1000);
      }
  };
}

export function publish(data: any) {
  if (!ws) return false;
  try {
    ws.send(JSON.stringify(data));
    return true;
  } catch (e) {
    return false;
  }
}

export function disconnect() {
  if (reconnectTimer) window.clearTimeout(reconnectTimer);
  if (ws) {
    ws.close();
    ws = null;
  }
  wsStatus.setWsStatus('disconnected');
}

export default { subscribe, publish, disconnect };
