/**
 * WebSocket Connection Manager
 * 
 * Optimizes WebSocket connections by implementing:
 * - Connection pooling and reuse
 * - Exponential backoff for reconnections
 * - Proper cleanup and abort mechanisms
 * - Connection state management
 * - Message multiplexing for multiple consumers
 */

type WSMessageHandler = (data: any) => void;
type WSStateHandler = (connected: boolean) => void;

interface WSSubscription {
  id: string;
  messageHandler?: WSMessageHandler;
  stateHandler?: WSStateHandler;
}

export class OptimizedWebSocketManager {
  private static instance: OptimizedWebSocketManager;
  private connections: Map<string, WebSocket> = new Map();
  private subscriptions: Map<string, WSSubscription[]> = new Map();
  private reconnectAttempts: Map<string, number> = new Map();
  private reconnectTimers: Map<string, number> = new Map();
  private baseReconnectDelay = 1000; // Start with 1 second
  private maxReconnectDelay = 30000; // Max 30 seconds
  private maxReconnectAttempts = 10;

  private constructor() {}

  static getInstance(): OptimizedWebSocketManager {
    if (!OptimizedWebSocketManager.instance) {
      OptimizedWebSocketManager.instance = new OptimizedWebSocketManager();
    }
    return OptimizedWebSocketManager.instance;
  }

  /**
   * Subscribe to a WebSocket connection with automatic management
   */
  subscribe(
    url: string,
    subscriberId: string,
    messageHandler?: WSMessageHandler,
    stateHandler?: WSStateHandler
  ): () => void {
    // Add subscription
    if (!this.subscriptions.has(url)) {
      this.subscriptions.set(url, []);
    }
    
    const subscription: WSSubscription = {
      id: subscriberId,
      messageHandler,
      stateHandler,
    };
    
    this.subscriptions.get(url)!.push(subscription);

    // Create connection if it doesn't exist
    this.ensureConnection(url);

    // Return unsubscribe function
    return () => this.unsubscribe(url, subscriberId);
  }

  private unsubscribe(url: string, subscriberId: string) {
    const subs = this.subscriptions.get(url);
    if (!subs) return;

    const filteredSubs = subs.filter(sub => sub.id !== subscriberId);
    
    if (filteredSubs.length === 0) {
      // No more subscribers, close connection
      this.closeConnection(url);
      this.subscriptions.delete(url);
    } else {
      this.subscriptions.set(url, filteredSubs);
    }
  }

  private ensureConnection(url: string) {
    if (this.connections.has(url)) {
      const ws = this.connections.get(url)!;
      if (ws.readyState === WebSocket.OPEN) {
        // Connection is already open, notify state handlers
        this.notifyStateHandlers(url, true);
        return;
      } else if (ws.readyState === WebSocket.CONNECTING) {
        // Connection is still connecting, wait for it
        return;
      }
    }

    this.createConnection(url);
  }

  private createConnection(url: string) {
    try {
      const ws = new WebSocket(url);
      this.connections.set(url, ws);

      ws.onopen = () => {
        console.log(`WebSocket connected to ${url}`);
        this.reconnectAttempts.set(url, 0); // Reset reconnect attempts on success
        this.notifyStateHandlers(url, true);
      };

      ws.onclose = (event) => {
        console.log(`WebSocket disconnected from ${url} (code: ${event.code})`);
        this.connections.delete(url);
        this.notifyStateHandlers(url, false);
        
        // Only attempt reconnection if there are still subscribers
        const subs = this.subscriptions.get(url);
        if (subs && subs.length > 0) {
          this.scheduleReconnect(url);
        }
      };

      ws.onerror = (error) => {
        console.warn(`WebSocket error for ${url}:`, error);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.notifyMessageHandlers(url, data);
        } catch (parseError) {
          console.warn(`Failed to parse WebSocket message from ${url}:`, parseError);
        }
      };

    } catch (error) {
      console.error(`Failed to create WebSocket connection to ${url}:`, error);
      this.scheduleReconnect(url);
    }
  }

  private scheduleReconnect(url: string) {
    const attempts = this.reconnectAttempts.get(url) || 0;
    
    if (attempts >= this.maxReconnectAttempts) {
      console.warn(`Max reconnection attempts (${this.maxReconnectAttempts}) reached for ${url}`);
      return;
    }

    // Exponential backoff with jitter
    const delay = Math.min(
      this.baseReconnectDelay * Math.pow(2, attempts),
      this.maxReconnectDelay
    );
    const jitter = Math.random() * 1000; // Add up to 1 second jitter
    const totalDelay = delay + jitter;

    console.log(`Scheduling reconnection to ${url} in ${Math.round(totalDelay)}ms (attempt ${attempts + 1})`);

    // Clear any existing timer
    const existingTimer = this.reconnectTimers.get(url);
    if (existingTimer) {
      window.clearTimeout(existingTimer);
    }

    const timer = window.setTimeout(() => {
      this.reconnectAttempts.set(url, attempts + 1);
      this.reconnectTimers.delete(url);
      this.createConnection(url);
    }, totalDelay);

    this.reconnectTimers.set(url, timer);
  }

  private notifyStateHandlers(url: string, connected: boolean) {
    const subs = this.subscriptions.get(url);
    if (!subs) return;

    subs.forEach(sub => {
      try {
        sub.stateHandler?.(connected);
      } catch (error) {
        console.error(`Error in state handler for ${url}:`, error);
      }
    });
  }

  private notifyMessageHandlers(url: string, data: any) {
    const subs = this.subscriptions.get(url);
    if (!subs) return;

    subs.forEach(sub => {
      try {
        sub.messageHandler?.(data);
      } catch (error) {
        console.error(`Error in message handler for ${url}:`, error);
      }
    });
  }

  private closeConnection(url: string) {
    const ws = this.connections.get(url);
    if (ws) {
      ws.close();
      this.connections.delete(url);
    }

    // Clear reconnection timer
    const timer = this.reconnectTimers.get(url);
    if (timer) {
      window.clearTimeout(timer);
      this.reconnectTimers.delete(url);
    }

    // Reset reconnect attempts
    this.reconnectAttempts.delete(url);
  }

  /**
   * Send message through a WebSocket connection
   */
  send(url: string, data: any): boolean {
    const ws = this.connections.get(url);
    if (ws && ws.readyState === WebSocket.OPEN) {
      try {
        ws.send(JSON.stringify(data));
        return true;
      } catch (error) {
        console.error(`Failed to send message to ${url}:`, error);
        return false;
      }
    }
    return false;
  }

  /**
   * Get connection state
   */
  isConnected(url: string): boolean {
    const ws = this.connections.get(url);
    return ws ? ws.readyState === WebSocket.OPEN : false;
  }

  /**
   * Force close all connections (useful for cleanup)
   */
  closeAll() {
    const urls = Array.from(this.connections.keys());
    urls.forEach(url => this.closeConnection(url));
    this.subscriptions.clear();
  }

  /**
   * Get connection statistics
   */
  getStats(): Record<string, any> {
    const stats: Record<string, any> = {};
    
    for (const [url] of this.connections) {
      const ws = this.connections.get(url);
      const subs = this.subscriptions.get(url) || [];
      const attempts = this.reconnectAttempts.get(url) || 0;
      
      stats[url] = {
        state: ws ? ['CONNECTING', 'OPEN', 'CLOSING', 'CLOSED'][ws.readyState] : 'DISCONNECTED',
        subscribers: subs.length,
        reconnectAttempts: attempts,
        hasReconnectTimer: this.reconnectTimers.has(url),
      };
    }
    
    return stats;
  }
}

// Export singleton instance
export const wsManager = OptimizedWebSocketManager.getInstance();

import { useEffect, useRef } from 'react';

// React hook for easy usage
export function useOptimizedWebSocket(
  url: string,
  subscriberId: string,
  messageHandler?: WSMessageHandler,
  stateHandler?: WSStateHandler
) {
  const unsubscribeRef = useRef<(() => void) | null>(null);
  
  useEffect(() => {
    // Subscribe to WebSocket
    const unsubscribe = wsManager.subscribe(url, subscriberId, messageHandler, stateHandler);
    unsubscribeRef.current = unsubscribe;
    
    // Cleanup on unmount
    return () => {
      unsubscribe();
      unsubscribeRef.current = null;
    };
  }, [url, subscriberId]); // Only re-subscribe if URL or ID changes
  
  return {
    send: (data: any) => wsManager.send(url, data),
    isConnected: () => wsManager.isConnected(url),
  };
}