export interface SSEEvent {
  event: string;
  data: string;
}

export type SSEHandler = (event: SSEEvent) => void;

export type Disconnect = () => void;

export function connectSSE(
  url: string,
  token: string,
  onEvent: SSEHandler,
  onError?: (err: unknown) => void,
): Disconnect {
  let ws: WebSocket | null = null;
  let retryDelay = 1000;
  let retryTimer: ReturnType<typeof setTimeout> | null = null;
  let closed = false;

  function connect() {
    if (closed) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}${url}`;

    ws = new WebSocket(wsUrl, ['bearer', token]);

    ws.onopen = () => {
      retryDelay = 1000;
    };

    ws.onmessage = (e: MessageEvent) => {
      try {
        const msg = JSON.parse(e.data);
        if (msg.type && msg.data !== undefined) {
          onEvent({ event: msg.type, data: JSON.stringify(msg.data) });
        }
      } catch {
        // ignore malformed messages
      }
    };

    ws.onerror = (err) => {
      if (!closed) {
        onError?.(err);
      }
    };

    ws.onclose = () => {
      ws = null;
      if (!closed) {
        retryTimer = setTimeout(() => {
          retryDelay = Math.min(retryDelay * 2, 30000);
          connect();
        }, retryDelay);
      }
    };
  }

  connect();

  return () => {
    closed = true;
    if (retryTimer !== null) {
      clearTimeout(retryTimer);
      retryTimer = null;
    }
    ws?.close();
    ws = null;
  };
}
