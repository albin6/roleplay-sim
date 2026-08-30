import { useEffect, useRef, useCallback, useState } from "react";
import { useAuthStore } from "../stores/authStore";
import { useRoomStore } from "../stores/roomStore";
import { WSEnvelope } from "../lib/types";

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/v1/ws";

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const token = useAuthStore((s) => s.token);
  const handleWSEvent = useRoomStore((s) => s.handleWSEvent);
  const seqRef = useRef(0);
  const listenersRef = useRef<Map<string, Set<(payload: any) => void>>>(new Map());

  const send = useCallback((event: string, payload: any = {}) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket not open, cannot send event:", event);
      return;
    }
    seqRef.current += 1;
    const envelope: WSEnvelope = {
      event,
      payload,
      timestamp: new Date().toISOString(),
      seq: seqRef.current,
    };
    wsRef.current.send(JSON.stringify(envelope));
  }, []);

  const sendBinary = useCallback((data: Blob | ArrayBuffer) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket not open, cannot send binary data");
      return;
    }
    wsRef.current.send(data);
  }, []);

  const on = useCallback((event: string, callback: (payload: any) => void) => {
    if (!listenersRef.current.has(event)) {
      listenersRef.current.set(event, new Set());
    }
    listenersRef.current.get(event)!.add(callback);

    return () => {
      listenersRef.current.get(event)?.delete(callback);
    };
  }, []);

  useEffect(() => {
    if (!token) return;

    let reconnectTimer: NodeJS.Timeout;
    let isSubscribed = true;

    const connect = () => {
      try {
        const url = `${WS_URL}?token=${encodeURIComponent(token)}`;
        const socket = new WebSocket(url);
        wsRef.current = socket;

        socket.onopen = () => {
          if (!isSubscribed) return;
          console.log("WebSocket connected");
          setIsConnected(true);
        };

        socket.onmessage = (event) => {
          try {
            const data: WSEnvelope = JSON.parse(event.data);
            handleWSEvent(data);

            const listeners = listenersRef.current.get(data.event);
            if (listeners) {
              listeners.forEach((cb) => cb(data.payload));
            }
          } catch (e) {
            console.error("Failed to parse WS message:", e);
          }
        };

        socket.onclose = () => {
          if (!isSubscribed) return;
          console.log("WebSocket closed, reconnecting in 3s...");
          setIsConnected(false);
          reconnectTimer = setTimeout(connect, 3000);
        };

        socket.onerror = (err) => {
          console.error("WebSocket error:", err);
          socket.close();
        };
      } catch (err) {
        console.error("WebSocket connection failure:", err);
      }
    };

    connect();

    return () => {
      isSubscribed = false;
      clearTimeout(reconnectTimer);
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [token, handleWSEvent]);

  return { isConnected, send, sendBinary, on };
}