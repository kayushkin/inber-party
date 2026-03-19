import { useState, useEffect } from 'react';
import { useNetworkErrorHandler } from '../hooks/useErrorHandler';

// Configuration for API client
export interface ApiClientConfig {
  baseUrl?: string;
  timeout?: number;
  retries?: number;
  retryDelay?: number;
  offlineMode?: boolean;
}

// Error types
export class ApiError extends Error {
  public status?: number;
  public statusText?: string;
  public isNetworkError: boolean;
  
  constructor(
    message: string,
    status?: number,
    statusText?: string,
    isNetworkError: boolean = false
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.statusText = statusText;
    this.isNetworkError = isNetworkError;
  }
}

export class OfflineError extends Error {
  constructor() {
    super('Application is offline. Please check your connection.');
    this.name = 'OfflineError';
  }
}

// Network status tracker
class NetworkStatus {
  private isOnlineState = navigator.onLine;
  private listeners: Set<(isOnline: boolean) => void> = new Set();

  constructor() {
    window.addEventListener('online', this.handleOnline);
    window.addEventListener('offline', this.handleOffline);
  }

  private handleOnline = () => {
    this.isOnlineState = true;
    this.notifyListeners();
  };

  private handleOffline = () => {
    this.isOnlineState = false;
    this.notifyListeners();
  };

  private notifyListeners() {
    this.listeners.forEach(listener => listener(this.isOnlineState));
  }

  get isOnline() {
    return this.isOnlineState;
  }

  subscribe(callback: (isOnline: boolean) => void) {
    this.listeners.add(callback);
    return () => { 
      this.listeners.delete(callback);
    };
  }

  cleanup() {
    window.removeEventListener('online', this.handleOnline);
    window.removeEventListener('offline', this.handleOffline);
    this.listeners.clear();
  }
}

// Global network status instance
export const networkStatus = new NetworkStatus();

// Exponential backoff delay calculator
function getRetryDelay(attempt: number, baseDelay: number = 1000): number {
  // Exponential backoff: 1s, 2s, 4s, 8s, etc. with jitter
  const delay = Math.min(baseDelay * Math.pow(2, attempt), 30000);
  const jitter = Math.random() * 0.1 * delay;
  return delay + jitter;
}

// Check if error is retryable
function isRetryableError(error: Error | Response): boolean {
  if (error instanceof Response) {
    // Retry on server errors (5xx) and rate limiting (429)
    return error.status >= 500 || error.status === 429;
  }
  
  if (error instanceof ApiError) {
    return error.isNetworkError || (error.status && error.status >= 500) || error.status === 429;
  }
  
  // Network errors, timeouts
  return error.message.toLowerCase().includes('network') ||
         error.message.toLowerCase().includes('timeout') ||
         error.message.toLowerCase().includes('fetch');
}

// Main API client class
export class ApiClient {
  private config: Required<ApiClientConfig>;
  private abortControllers = new Map<string, AbortController>();

  constructor(config: ApiClientConfig = {}) {
    this.config = {
      baseUrl: config.baseUrl || import.meta.env.VITE_API_URL || '',
      timeout: config.timeout || 10000,
      retries: config.retries || 3,
      retryDelay: config.retryDelay || 1000,
      offlineMode: config.offlineMode ?? true
    };
  }

  private async delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  private createRequestId(): string {
    return Math.random().toString(36).substring(2) + Date.now().toString(36);
  }

  private getAbortController(requestId: string): AbortController {
    const controller = new AbortController();
    this.abortControllers.set(requestId, controller);
    return controller;
  }

  private cleanupAbortController(requestId: string): void {
    this.abortControllers.delete(requestId);
  }

  // Cancel specific request
  cancelRequest(requestId: string): void {
    const controller = this.abortControllers.get(requestId);
    if (controller) {
      controller.abort();
      this.cleanupAbortController(requestId);
    }
  }

  // Cancel all pending requests
  cancelAllRequests(): void {
    this.abortControllers.forEach(controller => controller.abort());
    this.abortControllers.clear();
  }

  private async makeRequest(
    url: string, 
    options: RequestInit = {},
    requestId: string,
    attempt: number = 0
  ): Promise<Response> {
    // Check offline mode
    if (this.config.offlineMode && !networkStatus.isOnline) {
      throw new OfflineError();
    }

    const controller = this.getAbortController(requestId);
    
    // Setup timeout
    const timeoutId = setTimeout(() => {
      controller.abort();
    }, this.config.timeout);

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal
      });

      clearTimeout(timeoutId);
      this.cleanupAbortController(requestId);

      // Handle HTTP errors
      if (!response.ok) {
        const error = new ApiError(
          `HTTP ${response.status}: ${response.statusText}`,
          response.status,
          response.statusText
        );

        // Retry if it's a retryable error and we haven't exceeded max retries
        if (attempt < this.config.retries && isRetryableError(response)) {
          console.warn(`Request failed (attempt ${attempt + 1}/${this.config.retries + 1}): ${error.message}`);
          await this.delay(getRetryDelay(attempt, this.config.retryDelay));
          return this.makeRequest(url, options, requestId, attempt + 1);
        }

        throw error;
      }

      return response;

    } catch (error) {
      clearTimeout(timeoutId);
      this.cleanupAbortController(requestId);

      if (error instanceof Error) {
        // Handle abort
        if (error.name === 'AbortError') {
          throw new ApiError('Request was cancelled', 0, 'Cancelled');
        }

        // Convert network errors
        const isNetworkError = 
          error.message.toLowerCase().includes('network') ||
          error.message.toLowerCase().includes('fetch') ||
          !navigator.onLine;

        const apiError = new ApiError(
          isNetworkError ? 'Network error - please check your connection' : error.message,
          0,
          error.message,
          isNetworkError
        );

        // Retry if it's retryable and we haven't exceeded max retries
        if (attempt < this.config.retries && isRetryableError(apiError)) {
          console.warn(`Request failed (attempt ${attempt + 1}/${this.config.retries + 1}): ${apiError.message}`);
          await this.delay(getRetryDelay(attempt, this.config.retryDelay));
          return this.makeRequest(url, options, requestId, attempt + 1);
        }

        throw apiError;
      }

      throw error;
    }
  }

  private getFullUrl(endpoint: string): string {
    if (endpoint.startsWith('http')) {
      return endpoint;
    }
    return `${this.config.baseUrl}${endpoint.startsWith('/') ? endpoint : '/' + endpoint}`;
  }

  // Generic request method
  async request<T>(
    endpoint: string,
    options: RequestInit = {},
    parser?: (response: Response) => Promise<T>
  ): Promise<{ data: T; requestId: string }> {
    const requestId = this.createRequestId();
    const url = this.getFullUrl(endpoint);

    const response = await this.makeRequest(url, options, requestId);
    
    let data: T;
    if (parser) {
      data = await parser(response);
    } else {
      data = await response.json();
    }

    return { data, requestId };
  }

  // Convenience methods
  async get<T>(endpoint: string, params?: Record<string, string>): Promise<{ data: T; requestId: string }> {
    const url = new URL(this.getFullUrl(endpoint));
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, value);
      });
    }

    return this.request<T>(url.toString());
  }

  async post<T>(endpoint: string, body?: unknown): Promise<{ data: T; requestId: string }> {
    return this.request<T>(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async put<T>(endpoint: string, body?: unknown): Promise<{ data: T; requestId: string }> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<{ data: T; requestId: string }> {
    return this.request<T>(endpoint, {
      method: 'DELETE',
    });
  }

  // Upload file with progress tracking
  async uploadFile<T>(
    endpoint: string,
    file: File,
    onProgress?: (progress: number) => void
  ): Promise<{ data: T; requestId: string }> {
    return new Promise((resolve, reject) => {
      const requestId = this.createRequestId();
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('file', file);

      if (onProgress) {
        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable) {
            const progress = (event.loaded / event.total) * 100;
            onProgress(progress);
          }
        });
      }

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const data = JSON.parse(xhr.responseText);
            resolve({ data, requestId });
          } catch {
            reject(new ApiError('Invalid JSON response', xhr.status));
          }
        } else {
          reject(new ApiError(`HTTP ${xhr.status}: ${xhr.statusText}`, xhr.status, xhr.statusText));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new ApiError('Upload failed', 0, 'Network Error', true));
      });

      xhr.addEventListener('timeout', () => {
        reject(new ApiError('Upload timeout', 0, 'Timeout'));
      });

      xhr.timeout = this.config.timeout;
      xhr.open('POST', this.getFullUrl(endpoint));
      xhr.send(formData);

      // Store reference for potential cancellation
      this.abortControllers.set(requestId, {
        abort: () => xhr.abort()
      } as AbortController);
    });
  }
}

// Default API client instance
export const apiClient = new ApiClient();

// React hook for API calls with error handling
export function useApiClient() {
  const errorHandler = useNetworkErrorHandler();

  const makeApiCall = async <T>(
    apiCall: () => Promise<{ data: T; requestId: string }>,
    errorMessage?: string
  ): Promise<{ data: T; requestId: string } | null> => {
    return errorHandler.withErrorHandling(async () => {
      return await apiCall();
    }, errorMessage);
  };

  return {
    ...errorHandler,
    apiClient,
    makeApiCall,
    isOnline: networkStatus.isOnline
  };
}

// Hook to monitor network status
export function useNetworkStatus() {
  const [isOnline, setIsOnline] = useState(networkStatus.isOnline);

  useEffect(() => {
    const unsubscribe = networkStatus.subscribe(setIsOnline);
    return () => unsubscribe();
  }, []);

  return isOnline;
}

// Export necessary types and instances
export default apiClient;