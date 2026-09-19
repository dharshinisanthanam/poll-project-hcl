import { useEffect, useRef, useState, useCallback } from 'react';

export function usePollLive(pollIdOrShareCode, onUpdate) {
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  const connect = useCallback(() => {
    if (!pollIdOrShareCode) return;

    // Determine WS protocol and host safely for any device
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const rawWsUrl = import.meta.env.VITE_WS_URL;
    let host = `${protocol}//${window.location.host}`;
    if (rawWsUrl && rawWsUrl.trim() !== '') {
      if (window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1' && rawWsUrl.includes('localhost')) {
        host = `${protocol}//${window.location.host}`;
      } else {
        host = rawWsUrl.trim();
      }
    }
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
