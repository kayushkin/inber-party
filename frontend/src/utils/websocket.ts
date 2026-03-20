/**
 * WebSocket Connection Manager
 * 
 * Optimizes WebSocket connections by implementing:
 * - Connection pooling and reuse
 * - Exponential backoff for reconnections
 * - Proper cleanup and abort mechanisms
 * - Connection state management
 * - Message multiplexing for multiple consumers
 * - Test environment optimizations
 */

type WSMessageHandler = (data: unknown) => void;
type WSStateHandler = (connected: boolean) => void;

// WebSocket connection statistics interface
interface WSConnectionStats {
  state: 'CONNECTING' | 'OPEN' | 'CLOSING' | 'CLOSED' | 'DISCONNECTED' | 'PENDING';
  subscribers: number;
  reconnectAttempts: number;
  hasReconnectTimer: boolean;
  hasConnectionTimer: boolean;
  isTestEnvironment: boolean;
}

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
  private connectionTimers: Map<string, number> = new Map();
  private baseReconnectDelay = 1000; // Start with 1 second
  private maxReconnectDelay = 30000; // Max 30 seconds
  private maxReconnectAttempts = 10;
  private isTestEnvironment = false;
  private testConnectionDelay = 2000; // 2 second delay in test environment

  private constructor() {
    // Detect test environment
    this.isTestEnvironment = this.detectTestEnvironment();
    
    // Adjust settings for test environment
    if (this.isTestEnvironment) {
      this.maxReconnectAttempts = 3; // Fewer reconnect attempts during tests
      this.maxReconnectDelay = 10000; // Shorter max delay
      console.log('🧪 WebSocket Manager: Test environment detected, using optimized settings');
    }
  }

  private detectTestEnvironment(): boolean {
    // Multiple ways to detect test environment
    return !!(
      // Playwright test runner
      '__playwright' in globalThis || 
      // Jest test environment
      '__jest' in globalThis ||
      // Environment variables (using import.meta.env for Vite)
      import.meta.env.VITE_NODE_ENV === 'test' ||
      import.meta.env.VITE_CI === 'true' ||
      // User agent detection for headless browsers
      (typeof navigator !== 'undefined' && (
        navigator.userAgent.includes('HeadlessChrome') ||
        navigator.userAgent.includes('Firefox') && navigator.webdriver
      )) ||
      // Location detection for test servers
      (typeof window !== 'undefined' && 
        window.location.hostname === 'localhost' && 
        window.location.port === '5173'
      )
    );
  }

  private async checkServerReadiness(): Promise<boolean> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 2000); // 2 second timeout
      
      const response = await fetch('/api/health', {
        method: 'GET',
        signal: controller.signal
      });
      
      clearTimeout(timeoutId);
      return response.ok;
    } catch (error) {
      console.log('🧪 Server readiness check failed:', error);
      return false;
    }
  }

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

    // Check if there's already a pending connection timer
    if (this.connectionTimers.has(url)) {
      return;
    }

    // In test environment, delay connection to allow backend to fully start
    if (this.isTestEnvironment) {
      console.log(`🧪 Delaying WebSocket connection to ${url} by ${this.testConnectionDelay}ms for test stability`);
      
      const timer = window.setTimeout(() => {
        this.connectionTimers.delete(url);
        this.createConnection(url);
      }, this.testConnectionDelay);
      
      this.connectionTimers.set(url, timer);
    } else {
      this.createConnection(url);
    }
  }

  private async createConnection(url: string) {
    // In test environment, check server readiness first
    if (this.isTestEnvironment) {
      const isReady = await this.checkServerReadiness();
      if (!isReady) {
        console.log('🧪 Backend not ready, delaying WebSocket connection');
        this.scheduleReconnect(url);
        return;
      }
    }

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

  private notifyMessageHandlers(url: string, data: unknown) {
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
    const reconnectTimer = this.reconnectTimers.get(url);
    if (reconnectTimer) {
      window.clearTimeout(reconnectTimer);
      this.reconnectTimers.delete(url);
    }

    // Clear connection timer (for delayed connections)
    const connectionTimer = this.connectionTimers.get(url);
    if (connectionTimer) {
      window.clearTimeout(connectionTimer);
      this.connectionTimers.delete(url);
    }

    // Reset reconnect attempts
    this.reconnectAttempts.delete(url);
  }

  /**
   * Send message through a WebSocket connection
   */
  send(url: string, data: unknown): boolean {
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
    
    // Extra cleanup for any remaining timers
    this.reconnectTimers.forEach(timer => window.clearTimeout(timer));
    this.reconnectTimers.clear();
    this.connectionTimers.forEach(timer => window.clearTimeout(timer));
    this.connectionTimers.clear();
    this.reconnectAttempts.clear();
  }

  /**
   * Get connection statistics
   */
  getStats(): Record<string, WSConnectionStats> {
    const stats: Record<string, WSConnectionStats> = {};
    
    for (const [url] of this.connections) {
      const ws = this.connections.get(url);
      const subs = this.subscriptions.get(url) || [];
      const attempts = this.reconnectAttempts.get(url) || 0;
      
      stats[url] = {
        state: ws ? (['CONNECTING', 'OPEN', 'CLOSING', 'CLOSED'][ws.readyState] as WSConnectionStats['state']) : 'DISCONNECTED',
        subscribers: subs.length,
        reconnectAttempts: attempts,
        hasReconnectTimer: this.reconnectTimers.has(url),
        hasConnectionTimer: this.connectionTimers.has(url),
        isTestEnvironment: this.isTestEnvironment,
      };
    }
    
    // Add info about pending subscriptions without active connections
    for (const [url, subs] of this.subscriptions) {
      if (!stats[url]) {
        stats[url] = {
          state: this.connectionTimers.has(url) ? 'PENDING' : 'DISCONNECTED',
          subscribers: subs.length,
          reconnectAttempts: this.reconnectAttempts.get(url) || 0,
          hasReconnectTimer: this.reconnectTimers.has(url),
          hasConnectionTimer: this.connectionTimers.has(url),
          isTestEnvironment: this.isTestEnvironment,
        };
      }
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
  }, [url, subscriberId, messageHandler, stateHandler]); // Re-subscribe if URL, ID, or handlers change
  
  return {
    send: (data: unknown) => wsManager.send(url, data),
    isConnected: () => wsManager.isConnected(url),
  };
}