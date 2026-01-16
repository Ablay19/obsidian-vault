import { metric } from './metrics';

interface TraceSpan {
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  operation: string;
  startTime: number;
  endTime?: number;
  durationMs?: number;
  tags?: Record<string, string>;
}

class DistributedTracer {
  private spans: Map<string, TraceSpan>;
  private currentTraceId: string | null;

  constructor() {
    this.spans = new Map();
    this.currentTraceId = null;
  }

  startTrace(operation: string, traceId?: string): TraceSpan {
    const id = crypto.randomUUID();
    const span: TraceSpan = {
      traceId: traceId || id,
      spanId: id,
      operation,
      startTime: Date.now(),
    };

    this.spans.set(id, span);
    this.currentTraceId = span.traceId;

    metric.histogram(`trace_${operation}_start_ms`, 0);

    return span;
  }

  endSpan(span: TraceSpan, error?: Error): void {
    span.endTime = Date.now();
    span.durationMs = span.endTime - span.startTime;

    metric.histogram(`trace_${span.operation}_duration_ms`, span.durationMs, {
      trace_id: span.traceId,
      span_id: span.spanId,
      success: !error ? 'true' : 'false',
    });

    if (error) {
      metric.increment('trace_error', 1, {
        operation: span.operation,
        error_type: error.name,
      });
    }

    this.spans.delete(span.spanId);
  }

  createChildSpan(parentSpan: TraceSpan, operation: string): TraceSpan {
    return this.startTrace(operation, parentSpan.traceId);
  }

  getTrace(traceId: string): TraceSpan[] {
    const trace: TraceSpan[] = [];

    for (const span of this.spans.values()) {
      if (span.traceId === traceId) {
        trace.push(span);
      }
    }

    return trace.sort((a, b) => a.startTime - b.startTime);
  }

  exportTrace(traceId: string): string {
    const spans = this.getTrace(traceId);

    const traceData = {
      trace_id: traceId,
      spans: spans.map(span => ({
        span_id: span.spanId,
        parent_span_id: span.parentSpanId,
        operation: span.operation,
        start_time: new Date(span.startTime).toISOString(),
        end_time: span.endTime ? new Date(span.endTime).toISOString() : null,
        duration_ms: span.durationMs,
        tags: span.tags,
      })),
    };

    return JSON.stringify(traceData, null, 2);
  }
}

const tracer = new DistributedTracer();

export const trace = {
  start: (operation: string, traceId?: string) => tracer.startTrace(operation, traceId),
  end: (span: TraceSpan, error?: Error) => tracer.endSpan(span, error),
  child: (parent: TraceSpan, operation: string) => tracer.createChildSpan(parent, operation),
  get: (traceId: string) => tracer.getTrace(traceId),
  export: (traceId: string) => tracer.exportTrace(traceId),
  currentTraceId: () => tracer['currentTraceId'],
};
