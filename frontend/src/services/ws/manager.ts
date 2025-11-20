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
  // Prefer deriving the WS endpoint from the browser location when available.
  // This makes the built SPA portable: opening the app on another host will
  // automatically connect back to the same origin for WebSocket upgrades.
  try {
    if (typeof window !== 'undefined' && window.location && window.location.host) {
      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
      return `${proto}://${window.location.host}/ws`;
    }
  } catch (e) {
    // fallthrough to API_BASE/config
  }

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
  const url = `${baseWs}`;
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
    // Send initial auth message instead of placing token in URL. Server will
    // respond with an auth_ok or close the connection on failure.
    try {
      ws?.send(JSON.stringify({ type: 'auth', token: token || '' }));
    } catch (e) {
      console.error('[wsManager] failed to send auth message', e);
      ws?.close();
    }
  };

  ws.onmessage = (ev) => {
    try {
      const parsed = JSON.parse(ev.data);
      // handle server auth_ok/auth_failed messages specially
      if (parsed && parsed.type === 'auth_ok') {
        // authenticated; ignore message
        return;
      }
      if (parsed && parsed.type === 'auth_failed') {
        console.warn('[wsManager] websocket auth failed, closing socket');
        ws?.close();
        return;
      }
      notifyAll(parsed);
    } catch (e) {
      // ignore non-json
    }
  };

  ws.onclose = (ev) => {
    ws = null;
    // Only reconnect on abnormal closures (not normal 1000/1001)
    // 1001 is "going away" which happens on tab switch - don't reconnect immediately
    if (ev && ev.code !== 1000 && ev.code !== 1001) {
      console.warn('[wsManager] connection closed abnormally, scheduling reconnect', ev.code);
      scheduleReconnect();
    } else {
      console.log('[wsManager] connection closed normally', ev.code);
      wsStatus.setWsStatus('disconnected');
      // Auto-reconnect after short delay for normal closures (tab refocus)
      if (ev && ev.code === 1001) {
        setTimeout(() => {
          if (!ws) connect();
        }, 1000);
      }
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
    // Keep socket alive even with no subscribers - don't close it
    // The backend sends pings every 30s to keep connection alive
  };
}

export function getSubscriberCount() {
  return subscribers.size;
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

export default { subscribe, publish, disconnect, getSubscriberCount };
