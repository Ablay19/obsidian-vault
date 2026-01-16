interface MetricData {
  name: string;
  value: number;
  timestamp: string;
  tags?: Record<string, string>;
}

class MetricsCollector {
  private metrics: Map<string, { sum: number; count: number; values: number[] }>;

  constructor() {
    this.metrics = new Map();
  }

  increment(name: string, value: number = 1, tags?: Record<string, string>): void {
    const metric = this.metrics.get(name) || { sum: 0, count: 0, values: [] };

    metric.sum += value;
    metric.count++;
    metric.values.push(value);

    this.metrics.set(name, metric);

    this.emitMetric({
      name,
      value,
      timestamp: new Date().toISOString(),
      tags,
    });
  }

  gauge(name: string, value: number, tags?: Record<string, string>): void {
    this.metrics.set(name, { sum: value, count: 1, values: [value] });

    this.emitMetric({
      name,
      value,
      timestamp: new Date().toISOString(),
      tags,
    });
  }

  histogram(name: string, value: number, tags?: Record<string, string>): void {
    const metric = this.metrics.get(name) || { sum: 0, count: 0, values: [] };

    metric.sum += value;
    metric.count++;
    metric.values.push(value);

    this.metrics.set(name, metric);

    this.emitMetric({
      name: `${name}_sum`,
      value: metric.sum,
      timestamp: new Date().toISOString(),
      tags,
    });

    this.emitMetric({
      name: `${name}_count`,
      value: metric.count,
      timestamp: new Date().toISOString(),
      tags,
    });

    this.emitMetric({
      name,
      value,
      timestamp: new Date().toISOString(),
      tags,
    });
  }

  timing(name: string, durationMs: number, tags?: Record<string, string>): void {
    this.histogram(name, durationMs, { ...tags, unit: 'ms' });
  }

  private emitMetric(data: MetricData): void {
    if (typeof console !== 'undefined') {
      const output = JSON.stringify(data);
      console.log(`METRIC: ${output}`);
    }
  }

  getMetrics(name: string): { sum: number; count: number; avg: number; min: number; max: number } | null {
    const metric = this.metrics.get(name);

    if (!metric || metric.values.length === 0) {
      return null;
    }

    const avg = metric.sum / metric.count;
    const min = Math.min(...metric.values);
    const max = Math.max(...metric.values);

    return { sum: metric.sum, count: metric.count, avg, min, max };
  }

  reset(): void {
    this.metrics.clear();
  }
}

const metrics = new MetricsCollector();

export const metric = {
  increment: (name: string, value?: number, tags?: Record<string, string>) => metrics.increment(name, value, tags),
  gauge: (name: string, value: number, tags?: Record<string, string>) => metrics.gauge(name, value, tags),
  histogram: (name: string, value: number, tags?: Record<string, string>) => metrics.histogram(name, value, tags),
  timing: (name: string, durationMs: number, tags?: Record<string, string>) => metrics.timing(name, durationMs, tags),
  get: (name: string) => metrics.getMetrics(name),
  reset: () => metrics.reset(),
};

export const predefinedMetrics = {
  problems_received: 'problems_received',
  ai_requests: 'ai_requests',
  ai_success: 'ai_success',
  ai_failure: 'ai_failure',
  render_requests: 'render_requests',
  render_success: 'render_success',
  render_failure: 'render_failure',
  video_delivery: 'video_delivery',
  telegram_messages_sent: 'telegram_messages_sent',
  error_occurred: 'error_occurred',
  request_duration_ms: 'request_duration_ms',
};
