type WsState = 'disconnected' | 'connecting' | 'connected' | 'error' | 'reconnecting';

type Listener = (s: WsState, info?: string) => void;

const listeners = new Set<Listener>();
let state: WsState = 'disconnected';
let info: string | undefined = undefined;

export function setWsStatus(s: WsState, details?: string) {
  state = s;
  info = details;
  listeners.forEach((l) => {
    try {
      l(state, info);
    } catch (e) {
      // swallow listener errors
      // console.error('wsStatus listener error', e)
    }
  });
}

export function getWsStatus() {
  return { state, info };
}

export function subscribeWsStatus(l: Listener) {
  listeners.add(l);
  return () => listeners.delete(l);
}

export default { setWsStatus, getWsStatus, subscribeWsStatus };
