interface Env {
  LOG_LEVEL: string;
  API_GATEWAY_URL: string;
}

interface APIResponse {
  status: string;
  data: unknown;
  message: string;
}

function createLogger(component: string, level: string) {
  return {
    info: (msg: string, data?: Record<string, unknown>) => {
      console.log(JSON.stringify({
        level: 'info',
        message: msg,
        component,
        ...data,
      }));
    },
    error: (msg: string, error?: Error, data?: Record<string, unknown>) => {
      console.log(JSON.stringify({
        level: 'error',
        message: msg,
        component,
        error: error?.message,
        stack: error?.stack,
        ...data,
      }));
    },
    debug: (msg: string, data?: Record<string, unknown>) => {
      console.log(JSON.stringify({
        level: 'debug',
        message: msg,
        component,
        ...data,
      }));
    },
  };
}

const logger = createLogger('ai-worker', 'info');

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    const startTime = Date.now();

    logger.info('Incoming request', {
      method: request.method,
      url: url.pathname,
    });

    try {
      if (url.pathname === '/health') {
        return handleHealth();
      }

      if (url.pathname.startsWith('/api/')) {
        return handleAPI(request, env, url);
      }

      return new Response('Not Found', { status: 404 });
    } catch (error) {
      logger.error('Request failed', error as Error, {
        method: request.method,
        url: url.pathname,
      });

      return new Response(
        JSON.stringify({
          status: 'error',
          message: (error as Error).message,
        } as APIResponse),
        { status: 500, headers: { 'Content-Type': 'application/json' } }
      );
    } finally {
      const duration = Date.now() - startTime;
      logger.debug('Request completed', {
        method: request.method,
        url: url.pathname,
        duration_ms: duration,
      });
    }
  },
};

function handleHealth(): Response {
  const response: APIResponse = {
    status: 'ok',
    data: {
      service: 'ai-worker',
      version: '1.0.0',
      timestamp: new Date().toISOString(),
    },
    message: 'Worker is healthy',
  };

  return new Response(JSON.stringify(response), {
    headers: { 'Content-Type': 'application/json' },
  });
}

async function handleAPI(request: Request, env: Env, url: URL): Promise<Response> {
  const apiGatewayUrl = env.API_GATEWAY_URL || 'http://localhost:8080';

  if (request.method === 'GET' && url.pathname === '/api/v1/workers') {
    return handleListWorkers(apiGatewayUrl);
  }

  if (request.method === 'GET' && url.pathname.startsWith('/api/v1/workers/')) {
    const workerId = url.pathname.split('/').pop();
    return handleGetWorker(apiGatewayUrl, workerId!);
  }

  return new Response('Not Found', { status: 404 });
}

async function handleListWorkers(apiGatewayUrl: string): Promise<Response> {
  try {
    const response = await fetch(`${apiGatewayUrl}/api/v1/workers`);
    const data = await response.json();
    return new Response(JSON.stringify(data), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    logger.error('Failed to list workers', error as Error);
    return new Response(
      JSON.stringify({
        status: 'error',
        message: 'Failed to fetch workers',
      } as APIResponse),
      { status: 500, headers: { 'Content-Type': 'application/json' } }
    );
  }
}

async function handleGetWorker(apiGatewayUrl: string, workerId: string): Promise<Response> {
  try {
    const response = await fetch(`${apiGatewayUrl}/api/v1/workers/${workerId}`);
    const data = await response.json();
    return new Response(JSON.stringify(data), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    logger.error('Failed to get worker', error as Error, { workerId });
    return new Response(
      JSON.stringify({
        status: 'error',
        message: 'Failed to fetch worker',
      } as APIResponse),
      { status: 500, headers: { 'Content-Type': 'application/json' } }
    );
  }
}
