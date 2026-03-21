/**
 * Test WebSocket Blocker
 * 
 * This module completely replaces all WebSocket functionality during tests
 * with no-op implementations to eliminate any possibility of connection churn.
 */

export class TestWebSocketBlocker {
  private static instance: TestWebSocketBlocker;
  private isTestEnvironment: boolean;

  private constructor() {
    this.isTestEnvironment = this.detectTestEnvironment();
    
    if (this.isTestEnvironment) {
      console.log('🧪 TEST WEBSOCKET BLOCKER: All WebSocket operations will be mocked');
      this.installGlobalBlocks();
    }
  }

  static getInstance(): TestWebSocketBlocker {
    if (!TestWebSocketBlocker.instance) {
      TestWebSocketBlocker.instance = new TestWebSocketBlocker();
    }
    return TestWebSocketBlocker.instance;
  }

  private detectTestEnvironment(): boolean {
    return !!(
      '__playwright' in globalThis ||
      '__jest' in globalThis ||
      (typeof navigator !== 'undefined' && (
        navigator.userAgent.includes('HeadlessChrome') ||
        navigator.userAgent.includes('Playwright') ||
        (navigator.userAgent.includes('Firefox') && navigator.webdriver)
      )) ||
      import.meta.env.VITE_NODE_ENV === 'test' ||
      import.meta.env.VITE_CI === 'true' ||
      (typeof window !== 'undefined' && window.location.hostname === 'localhost')
    );
  }

  private installGlobalBlocks() {
    if (typeof window === 'undefined') return;

    // Set global flags to indicate WebSocket blocking is active
    (window as any).__TEST_WEBSOCKET_BLOCKER_ACTIVE__ = true;
    (window as any).__WEBSOCKET_ABSOLUTE_LOCK__ = true;
    
    // Override WebSocket constructor to prevent any real connections
    // const OriginalWebSocket = window.WebSocket; // Keep reference if needed for restoration
    
    (window as any).WebSocket = class MockWebSocket {
      url: string;
      readyState: number = 1; // OPEN
      onopen: ((event: Event) => void) | null = null;
      onclose: ((event: CloseEvent) => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;

      constructor(url: string) {
        this.url = url;
        console.log(`🧪 MOCK WEBSOCKET: Blocked WebSocket creation to ${url}`);
        
        // Immediately mock connection success
        setTimeout(() => {
          this.onopen?.({} as Event);
        }, 0);
      }

      send(data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
        console.log(`🧪 MOCK WEBSOCKET: Blocked send to ${this.url}:`, data);
      }

      close(code?: number, reason?: string): void {
        console.log(`🧪 MOCK WEBSOCKET: Mock close for ${this.url}`, { code, reason });
        this.readyState = 3; // CLOSED
        setTimeout(() => {
          this.onclose?.({} as CloseEvent);
        }, 0);
      }

      static CONNECTING = 0;
      static OPEN = 1;
      static CLOSING = 2;
      static CLOSED = 3;
    };

    // Also set static properties on the mock
    (window as any).WebSocket.CONNECTING = 0;
    (window as any).WebSocket.OPEN = 1;
    (window as any).WebSocket.CLOSING = 2;
    (window as any).WebSocket.CLOSED = 3;

    console.log('🧪 GLOBAL WEBSOCKET OVERRIDE: All WebSocket connections are now mocked');
  }

  /**
   * Check if WebSocket operations should be blocked
   */
  shouldBlock(): boolean {
    return this.isTestEnvironment;
  }

  /**
   * Get mock WebSocket implementation for tests
   */
  getMockWebSocket() {
    if (!this.isTestEnvironment) return null;

    return {
      send: (data: unknown) => {
        console.log('🧪 MOCK: WebSocket send blocked', data);
        return true;
      },
      isConnected: () => {
        console.log('🧪 MOCK: WebSocket always connected');
        return true;
      },
      subscribe: (url: string, id: string, _messageHandler?: any, stateHandler?: any) => {
        console.log(`🧪 MOCK: WebSocket subscribe blocked - ${id} → ${url}`);
        
        // Immediately call state handler with connected state
        if (stateHandler) {
          setTimeout(() => stateHandler(true), 0);
        }

        // Return no-op unsubscribe function
        return () => {
          console.log(`🧪 MOCK: WebSocket unsubscribe blocked - ${id} from ${url}`);
        };
      }
    };
  }

  /**
   * Restore original WebSocket (for cleanup)
   */
  restore() {
    if (typeof window !== 'undefined' && this.isTestEnvironment) {
      console.log('🧪 TEST WEBSOCKET BLOCKER: Cleanup requested (but keeping mocks for test stability)');
      // We don't actually restore WebSocket during tests to maintain stability
    }
  }
}

// Initialize immediately if in test environment
const testBlocker = TestWebSocketBlocker.getInstance();

export default testBlocker;