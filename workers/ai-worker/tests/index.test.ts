import { describe, it, expect, beforeEach } from 'vitest';

interface Env {
  LOG_LEVEL: string;
  API_GATEWAY_URL: string;
}

const makeWorker = () => ({
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === '/health') {
      return new Response(
        JSON.stringify({
          status: 'ok',
          data: { service: 'ai-worker', version: '1.0.0' },
          message: 'Worker is healthy',
        }),
        { headers: { 'Content-Type': 'application/json' } }
      );
    }

    return new Response('Not Found', { status: 404 });
  },
});

describe('AI Worker', () => {
  let worker: { fetch: (request: Request, env: Env, ctx: ExecutionContext) => Promise<Response> };

  beforeEach(() => {
    worker = makeWorker();
  });

  describe('Health endpoint', () => {
    it('should return health status', async () => {
      const request = new Request('http://localhost/health');
      const env = { LOG_LEVEL: 'info', API_GATEWAY_URL: 'http://localhost:8080' };
      const ctx = {} as ExecutionContext;

      const response = await worker.fetch(request, env, ctx);

      expect(response.status).toBe(200);

      const data = await response.json();
      expect(data.status).toBe('ok');
      expect(data.data.service).toBe('ai-worker');
    });
  });

  describe('Not found', () => {
    it('should return 404 for unknown routes', async () => {
      const request = new Request('http://localhost/unknown');
      const env = { LOG_LEVEL: 'info', API_GATEWAY_URL: 'http://localhost:8080' };
      const ctx = {} as ExecutionContext;

      const response = await worker.fetch(request, env, ctx);

      expect(response.status).toBe(404);
    });
  });
});