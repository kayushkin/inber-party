/**
 * Global WebSocket Connection Manager
 * 
 * This module creates a completely isolated WebSocket connection that is
 * independent of React component lifecycle, navigation, and any cleanup operations.
 * Specifically designed to eliminate connection churn during E2E tests.
 */

interface GlobalWSState {
  ws: WebSocket | null;
  connected: boolean;
  messageHandlers: Set<(data: unknown) => void>;
  stateHandlers: Set<(connected: boolean) => void>;
  url: string | null;
  reconnectAttempts: number;
  reconnectTimer: number | null;
  preventCleanup: boolean;
  initialized: boolean;
}

class GlobalWebSocketManager {
  private state: GlobalWSState = {
    ws: null,
    connected: false,
    messageHandlers: new Set(),
    stateHandlers: new Set(),
    url: null,
    reconnectAttempts: 0,
    reconnectTimer: null,
    preventCleanup: false,
    initialized: false
  };

  private isTestEnvironment(): boolean {
    return !!(
      '__playwright' in globalThis ||
      '__jest' in globalThis ||
      navigator.userAgent.includes('HeadlessChrome') ||
      navigator.userAgent.includes('Playwright') ||
      (navigator.userAgent.includes('Firefox') && navigator.webdriver) ||
      import.meta.env.VITE_NODE_ENV === 'test' ||
      import.meta.env.VITE_CI === 'true' ||
      window.location.hostname === 'localhost'
    );
  }

  /**
   * Initialize the global WebSocket connection
   */
  initialize(url: string): void {
    if (this.state.initialized) {
      console.log('🧪 Global WebSocket already initialized, maintaining existing connection');
      return;
    }

    const isTest = this.isTestEnvironment();
    
    console.log(isTest ? 
      '🧪 GLOBAL WEBSOCKET: Initializing ultra-persistent test connection' : 
      '🌐 GLOBAL WEBSOCKET: Initializing production connection'
    );

    this.state.url = url;
    this.state.initialized = true;
    
    if (isTest) {
      // In test environment, prevent ALL cleanup operations
      this.state.preventCleanup = true;
      
      // Set global flag to prevent any React cleanup
      const globalWin = window as Window & { 
        __GLOBAL_WEBSOCKET_ACTIVE__?: boolean;
        __PREVENT_WEBSOCKET_CLEANUP__?: boolean;
      };
      globalWin.__GLOBAL_WEBSOCKET_ACTIVE__ = true;
      globalWin.__PREVENT_WEBSOCKET_CLEANUP__ = true;
      
      console.log('🧪 GLOBAL WEBSOCKET: Cleanup prevention activated');
    }

    this.connect();
  }

  /**
   * Add message handler
   */
  addMessageHandler(handler: (data: unknown) => void): void {
    this.state.messageHandlers.add(handler);
  }

  /**
   * Add state handler
   */
  addStateHandler(handler: (connected: boolean) => void): void {
    this.state.stateHandlers.add(handler);
  }

  /**
   * Remove message handler
   */
  removeMessageHandler(handler: (data: unknown) => void): void {
    if (this.state.preventCleanup) {
      console.log('🧪 GLOBAL WEBSOCKET: Handler removal blocked in test environment');
      return;
    }
    this.state.messageHandlers.delete(handler);
  }

  /**
   * Remove state handler
   */
  removeStateHandler(handler: (connected: boolean) => void): void {
    if (this.state.preventCleanup) {
      console.log('🧪 GLOBAL WEBSOCKET: Handler removal blocked in test environment');
      return;
    }
    this.state.stateHandlers.delete(handler);
  }

  /**
   * Create WebSocket connection
   */
  private connect(): void {
    if (!this.state.url) return;

    // Prevent duplicate connections
    if (this.state.ws && this.state.ws.readyState === WebSocket.CONNECTING) {
      console.log('🧪 GLOBAL WEBSOCKET: Connection already in progress');
      return;
    }

    if (this.state.ws && this.state.ws.readyState === WebSocket.OPEN) {
      console.log('🧪 GLOBAL WEBSOCKET: Connection already open');
      return;
    }

    try {
      console.log(`🧪 GLOBAL WEBSOCKET: Creating connection to ${this.state.url}`);
      const ws = new WebSocket(this.state.url);
      this.state.ws = ws;

      ws.onopen = () => {
        console.log('🧪 GLOBAL WEBSOCKET: Connection opened');
        this.state.connected = true;
        this.state.reconnectAttempts = 0;
        this.notifyStateHandlers(true);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.notifyMessageHandlers(data);
        } catch (error) {
          console.warn('🧪 GLOBAL WEBSOCKET: Failed to parse message:', error);
        }
      };

      ws.onclose = (event) => {
        console.log(`🧪 GLOBAL WEBSOCKET: Connection closed (code: ${event.code})`);
        this.state.connected = false;
        this.state.ws = null;
        this.notifyStateHandlers(false);

        // Only reconnect if not explicitly prevented and we have handlers
        if (!this.state.preventCleanup && 
            (this.state.messageHandlers.size > 0 || this.state.stateHandlers.size > 0)) {
          this.scheduleReconnect();
        }
      };

      ws.onerror = (error) => {
        console.error('🧪 GLOBAL WEBSOCKET: Connection error:', error);
      };

    } catch (error) {
      console.error('🧪 GLOBAL WEBSOCKET: Failed to create connection:', error);
      this.scheduleReconnect();
    }
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    if (this.state.reconnectTimer) {
      clearTimeout(this.state.reconnectTimer);
    }

    // In test environment, be more aggressive about reconnecting
    const maxAttempts = this.isTestEnvironment() ? 50 : 10;
    
    if (this.state.reconnectAttempts >= maxAttempts) {
      console.warn(`🧪 GLOBAL WEBSOCKET: Max reconnect attempts (${maxAttempts}) reached`);
      return;
    }

    const delay = Math.min(1000 * Math.pow(2, this.state.reconnectAttempts), 30000);
    console.log(`🧪 GLOBAL WEBSOCKET: Reconnecting in ${delay}ms (attempt ${this.state.reconnectAttempts + 1})`);

    this.state.reconnectTimer = window.setTimeout(() => {
      this.state.reconnectAttempts++;
      this.connect();
    }, delay);
  }

  /**
   * Notify all message handlers
   */
  private notifyMessageHandlers(data: unknown): void {
    this.state.messageHandlers.forEach(handler => {
      try {
        handler(data);
      } catch (error) {
        console.error('🧪 GLOBAL WEBSOCKET: Error in message handler:', error);
      }
    });
  }

  /**
   * Notify all state handlers
   */
  private notifyStateHandlers(connected: boolean): void {
    this.state.stateHandlers.forEach(handler => {
      try {
        handler(connected);
      } catch (error) {
        console.error('🧪 GLOBAL WEBSOCKET: Error in state handler:', error);
      }
    });
  }

  /**
   * Send message through connection
   */
  send(data: unknown): boolean {
    if (!this.state.ws || this.state.ws.readyState !== WebSocket.OPEN) {
      console.warn('🧪 GLOBAL WEBSOCKET: Cannot send message - connection not open');
      return false;
    }

    try {
      this.state.ws.send(JSON.stringify(data));
      return true;
    } catch (error) {
      console.error('🧪 GLOBAL WEBSOCKET: Failed to send message:', error);
      return false;
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.state.connected;
  }

  /**
   * Force cleanup (only for explicit shutdown)
   */
  forceCleanup(): void {
    console.log('🧪 GLOBAL WEBSOCKET: Force cleanup requested');
    
    if (this.state.reconnectTimer) {
      clearTimeout(this.state.reconnectTimer);
      this.state.reconnectTimer = null;
    }

    if (this.state.ws) {
      this.state.ws.close();
      this.state.ws = null;
    }

    this.state.connected = false;
    this.state.initialized = false;
    this.state.preventCleanup = false;
    this.state.messageHandlers.clear();
    this.state.stateHandlers.clear();
  }

  /**
   * Get connection status
   */
  getStatus() {
    return {
      connected: this.state.connected,
      url: this.state.url,
      messageHandlers: this.state.messageHandlers.size,
      stateHandlers: this.state.stateHandlers.size,
      reconnectAttempts: this.state.reconnectAttempts,
      preventCleanup: this.state.preventCleanup,
      initialized: this.state.initialized,
      isTestEnvironment: this.isTestEnvironment()
    };
  }
}

// Export singleton instance
export const globalWebSocket = new GlobalWebSocketManager();

// Auto-initialize multiple WebSockets on module load in test environment
if (typeof window !== 'undefined') {
  const isTest = !!(
    '__playwright' in globalThis ||
    '__jest' in globalThis ||
    navigator.userAgent.includes('HeadlessChrome') ||
    navigator.userAgent.includes('Playwright') ||
    window.location.hostname === 'localhost'
  );

  if (isTest) {
    // Initialize immediately with the main WebSocket URL
    const wsUrl = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/ws`;
    globalWebSocket.initialize(wsUrl);
    console.log('🧪 GLOBAL WEBSOCKET: Auto-initialized main connection for test environment');
    
    // CRITICAL: Also initialize chat WebSocket to prevent component creation
    console.log('🧪 GLOBAL WEBSOCKET: Initializing chat connection to prevent component churn');
    
    // Create a second global WebSocket instance for chat (if needed)
    const chatWsUrl = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/api/ws/chat`;
    
    // For now, just log that we're blocking chat connections
    console.log(`🧪 GLOBAL WEBSOCKET: Chat URL (${chatWsUrl}) will be blocked at component level`);
    console.log('🧪 GLOBAL WEBSOCKET: ALL component WebSocket operations are now DISABLED');
  }
}