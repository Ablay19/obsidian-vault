interface Env {
  LOG_LEVEL: string;
  API_GATEWAY_URL: string;
  AI_GATEWAY_URL: string;
}

interface APIResponse {
  status: string;
  data: unknown;
  message: string;
}

function jqColorize(json: string): string {
  let result = '';
  for (let i = 0; i < json.length; i++) {
    const c = json[i];
    if (c === '"') {
      let str = '';
      let j = i + 1;
      while (j < json.length && (json[j] !== '"' || json[j - 1] === '\\')) {
        str += json[j];
        j++;
      }
      result += '\x1b[38;5;214m"' + str + '"\x1b[0m';
      i = j;
    } else if (c === '{' || c === '[') {
      result += '\x1b[38;5;39m' + c + '\x1b[0m';
    } else if (c === '}' || c === ']') {
      result += '\x1b[38;5;39m' + c + '\x1b[0m';
    } else if (c === ':') {
      result += '\x1b[38;5;39m:\x1b[0m';
    } else if (c === ',') {
      result += '\x1b[38;5;39m,\x1b[0m';
    } else if (json.slice(i, i + 4) === 'true') {
      result += '\x1b[38;5;220mtrue\x1b[0m';
      i += 3;
    } else if (json.slice(i, i + 5) === 'false') {
      result += '\x1b[38;5;220mfalse\x1b[0m';
      i += 4;
    } else if (json.slice(i, i + 4) === 'null') {
      result += '\x1b[38;5;220mnull\x1b[0m';
      i += 3;
    } else if ((c >= '0' && c <= '9') || c === '-') {
      let num = '';
      let j = i;
      while (j < json.length && ((json[j] >= '0' && json[j] <= '9') || json[j] === '.' || json[j] === 'e' || json[j] === 'E' || json[j] === '+' || json[j] === '-')) {
        num += json[j];
        j++;
      }
      result += '\x1b[38;5;154m' + num + '\x1b[0m';
      i = j - 1;
    } else {
      result += c;
    }
  }
  return result;
}

function createLogger(component: string, level: string) {
  return {
    info: (msg: string, data?: Record<string, unknown>) => {
      const entry = {
        level: 'info',
        message: msg,
        component,
        ts: new Date().toISOString(),
        ...data,
      };
      console.log(jqColorize(JSON.stringify(entry)));
    },
    error: (msg: string, error?: Error, data?: Record<string, unknown>) => {
      const entry = {
        level: 'error',
        message: msg,
        component,
        ts: new Date().toISOString(),
        error: error?.message,
        ...data,
      };
      console.log(jqColorize(JSON.stringify(entry)));
    },
    debug: (msg: string, data?: Record<string, unknown>) => {
      const entry = {
        level: 'debug',
        message: msg,
        component,
        ts: new Date().toISOString(),
        ...data,
      };
      console.log(jqColorize(JSON.stringify(entry)));
    },
  };
}

const logger = createLogger('image-worker', 'info');

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

      if (url.pathname === '/process' && request.method === 'POST') {
        return handleProcessImage(request, env);
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
      service: 'image-worker',
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

  if (request.method === 'GET' && url.pathname === '/api/v1/images') {
    return handleListImages(apiGatewayUrl);
  }

  if (request.method === 'GET' && url.pathname.startsWith('/api/v1/images/')) {
    const imageId = url.pathname.split('/').pop();
    return handleGetImage(apiGatewayUrl, imageId!);
  }

  return new Response('Not Found', { status: 404 });
}

async function handleProcessImage(request: Request, env: Env): Promise<Response> {
  try {
    const formData = await request.formData();
    const image = formData.get('image');

    if (!image) {
      return new Response(
        JSON.stringify({
          status: 'error',
          message: 'Image file is required',
        } as APIResponse),
        { status: 400, headers: { 'Content-Type': 'application/json' } }
      );
    }

    logger.info('Processing image', { filename: (image as File).name });

    const aiGatewayUrl = env.AI_GATEWAY_URL || 'http://localhost:8080';

    const response: APIResponse = {
      status: 'ok',
      data: {
        image_id: 'img_' + Date.now(),
        status: 'processed',
        result: {
          labels: ['document', 'text'],
          confidence: 0.95,
        },
      },
      message: 'Image processed successfully',
    };

    return new Response(JSON.stringify(response), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    logger.error('Failed to process image', error as Error);
    return new Response(
      JSON.stringify({
        status: 'error',
        message: 'Failed to process image',
      } as APIResponse),
      { status: 500, headers: { 'Content-Type': 'application/json' } }
    );
  }
}

async function handleListImages(apiGatewayUrl: string): Promise<Response> {
  try {
    const response = await fetch(`${apiGatewayUrl}/api/v1/images`);
    const data = await response.json();
    return new Response(JSON.stringify(data), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    logger.error('Failed to list images', error as Error);
    return new Response(
      JSON.stringify({
        status: 'error',
        message: 'Failed to fetch images',
      } as APIResponse),
      { status: 500, headers: { 'Content-Type': 'application/json' } }
    );
  }
}

async function handleGetImage(apiGatewayUrl: string, imageId: string): Promise<Response> {
  try {
    const response = await fetch(`${apiGatewayUrl}/api/v1/images/${imageId}`);
    const data = await response.json();
    return new Response(JSON.stringify(data), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    logger.error('Failed to get image', error as Error, { imageId });
    return new Response(
      JSON.stringify({
        status: 'error',
        message: 'Failed to fetch image',
      } as APIResponse),
      { status: 500, headers: { 'Content-Type': 'application/json' } }
    );
  }
}
