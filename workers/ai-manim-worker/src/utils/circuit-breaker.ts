import { metric } from './metrics';

interface CircuitState {
  isOpen: boolean;
  failureCount: number;
  lastFailureTime: number;
  lastSuccessTime: number;
}

export class CircuitBreaker {
  private serviceName: string;
  private threshold: number;
  private timeoutMs: number;
  private state: CircuitState;
  private monitoring: boolean;

  constructor(serviceName: string, config?: { threshold?: number; timeoutMs?: number }) {
    this.serviceName = serviceName;
    this.threshold = config?.threshold || 5;
    this.timeoutMs = config?.timeoutMs || 60000;
    this.state = {
      isOpen: false,
      failureCount: 0,
      lastFailureTime: 0,
      lastSuccessTime: 0,
    };
    this.monitoring = false;
  }

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state.isOpen) {
      if (Date.now() - this.state.lastFailureTime < this.timeoutMs) {
        metric.increment('circuit_breaker_open', 1, { service: this.serviceName });
        throw new Error(`Circuit breaker open for ${this.serviceName}`);
      }

      this.reset();
    }

    this.monitoring = true;

    try {
      const result = await fn();

      this.recordSuccess();

      return result;
    } catch (error) {
      this.recordFailure();

      throw error;
    } finally {
      this.monitoring = false;
    }
  }

  private recordSuccess(): void {
    this.state.failureCount = 0;
    this.state.lastSuccessTime = Date.now();
    this.state.isOpen = false;

    metric.increment('circuit_breaker_success', 1, { service: this.serviceName });
  }

  private recordFailure(): void {
    this.state.failureCount++;
    this.state.lastFailureTime = Date.now();

    if (this.state.failureCount >= this.threshold) {
      this.state.isOpen = true;

      metric.increment('circuit_breaker_trip', 1, { service: this.serviceName });
    }

    metric.increment('circuit_breaker_failure', 1, { service: this.serviceName });
  }

  reset(): void {
    this.state = {
      isOpen: false,
      failureCount: 0,
      lastFailureTime: 0,
      lastSuccessTime: Date.now(),
    };

    metric.increment('circuit_breaker_reset', 1, { service: this.serviceName });
  }

  getState(): { isOpen: boolean; failureCount: number } {
    return {
      isOpen: this.state.isOpen,
      failureCount: this.state.failureCount,
    };
  }
}

export const createCircuitBreaker = (
  serviceName: string,
  config?: { threshold?: number; timeoutMs?: number }
): CircuitBreaker => {
  return new CircuitBreaker(serviceName, config);
};
