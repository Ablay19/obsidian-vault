export class RequestHandler {
  constructor() {
    // Constructor if needed
  }

  async parseRequest(request) {
    const url = new URL(request.url);
    const pathParts = url.pathname.split('/');
    const provider = pathParts[pathParts.length - 1];
    const prompt = await request.text();
    const clientIP = request.headers.get('cf-connecting-ip') || request.headers.get('x-forwarded-for');
    const userAgent = request.headers.get('user-agent');

    return {
      provider,
      prompt,
      clientIP,
      userAgent,
      url
    };
  }

  buildResponse(data, options = {}) {
    const { provider, cacheStatus = 'miss', responseTime, error, extraHeaders = {} } = options;

    const headers = {
      'content-type': 'application/json',
      'x-response-time': `${responseTime}ms`,
      ...extraHeaders
    };

    if (provider) {
      headers['x-ai-provider'] = provider;
    }

    if (cacheStatus) {
      headers['x-cache-status'] = cacheStatus;
    }

    if (error) {
      return new Response(JSON.stringify({ error: error.message }), {
        status: error.status || 500,
        headers
      });
    }

    return new Response(data, {
      status: 200,
      headers
    });
  }

  buildErrorResponse(error, responseTime) {
    return this.buildResponse(null, { error, responseTime });
  }
}