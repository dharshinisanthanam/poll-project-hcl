import { useEffect, useRef, useState, useCallback } from 'react';

export function usePollLive(pollIdOrShareCode, onUpdate) {
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  const connect = useCallback(() => {
    if (!pollIdOrShareCode) return;

    // Determine WS protocol and host safely for cloud deployments (Vercel -> Render)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const rawWsUrl = (import.meta.env.VITE_WS_URL || '').trim();
    const rawApiUrl = (import.meta.env.VITE_API_URL || '').trim();

    let host = '';

    if (rawWsUrl) {
      let cleaned = rawWsUrl.replace(/\/+$/, '');
      if (cleaned.startsWith('http://')) {
        cleaned = 'ws://' + cleaned.slice(7);
      } else if (cleaned.startsWith('https://')) {
        cleaned = 'wss://' + cleaned.slice(8);
      } else if (!cleaned.startsWith('ws://') && !cleaned.startsWith('wss://')) {
        cleaned = `${protocol}//${cleaned}`;
      }
      host = cleaned;
    } else if (rawApiUrl && rawApiUrl.startsWith('http')) {
      // Auto-derive WebSocket host from VITE_API_URL (e.g. https://poll.onrender.com/api -> wss://poll.onrender.com)
      try {
        const parsed = new URL(rawApiUrl);
        const wsProto = parsed.protocol === 'https:' ? 'wss:' : 'ws:';
        host = `${wsProto}//${parsed.host}`;
      } catch (e) {
        host = `${protocol}//${window.location.host}`;
      }
    } else {
      host = `${protocol}//${window.location.host}`;
    }

    host = host.replace(/\/+$/, '');
    const wsUrl = `${host}/ws/polls/${pollIdOrShareCode}`;


    try {
      const socket = new WebSocket(wsUrl);
      wsRef.current = socket;

      socket.onopen = () => {
        setIsConnected(true);
      };

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.event === 'VOTE_UPDATED' || data.event === 'POLL_STATUS_CHANGED') {
            if (onUpdate && data.payload) {
              onUpdate(data.payload);
            }
          }
        } catch (e) {
          console.error('Error parsing WebSocket message:', e);
        }
      };

      socket.onclose = () => {
        setIsConnected(false);
        // Attempt reconnection after 2 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, 2000);
      };

      socket.onerror = (err) => {
        console.warn('WebSocket connection error:', err);
        socket.close();
      };
    } catch (e) {
      console.error('Failed to establish WebSocket:', e);
    }
  }, [pollIdOrShareCode, onUpdate]);

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return { isConnected };
}
