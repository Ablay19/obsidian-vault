export class RequestHandler {
  constructor(options = {}) {
    this.analytics = options.analytics;
    this.tracer = options.tracer;
    this.profiler = options.profiler;
    this.errorHandler = options.errorHandler;

    // Performance tracking
    this.requestCount = 0;
    this.errorCount = 0;
    this.avgResponseTime = 0;
  }

  // Enhanced request processing with comprehensive error handling
  async processRequest(request, context) {
    const startTime = Date.now();
    this.requestCount++;

    try {
      // Start profiling if available
      if (this.profiler) {
        this.profiler.startProfile(`request-${this.requestCount}`);
      }

      // Parse and validate request
      const requestData = await this.parseRequestEnhanced(request);

      // Trace request
      if (this.tracer) {
        requestData.traceId = this.tracer.trace(request);
      }

      // Process the request
      const result = await this.processRequestLogic(requestData, context);

      // Calculate response time
      const responseTime = Date.now() - startTime;

      // Track analytics
      if (this.analytics) {
        this.analytics.trackRequest(responseTime, true, requestData.endpoint || 'unknown');
      }

      // Update average response time
      this.updateAverageResponseTime(responseTime);

      // Stop profiling
      if (this.profiler) {
        const profileResult = this.profiler.endProfile(`request-${this.requestCount}`);
        result.profile = profileResult;
      }

      return this.buildResponse(result, {
        provider: result.provider,
        cacheStatus: result.cacheStatus,
        responseTime,
        traceId: requestData.traceId
      });

    } catch (error) {
      this.errorCount++;
      const responseTime = Date.now() - startTime;

      // Track error analytics
      if (this.analytics) {
        this.analytics.trackRequest(responseTime, false, 'error');
      }

      // Handle error with recovery logic
      return await this.handleError(error, request, responseTime);
    }
  }

  async parseRequestEnhanced(request) {
    const url = new URL(request.url);
    const pathParts = url.pathname.split('/').filter(p => p);
    const method = request.method;

    // Enhanced URL parsing
    const endpoint = pathParts[pathParts.length - 1] || 'default';
    const provider = pathParts[pathParts.length - 2] || url.searchParams.get('provider');

    // Parse request body based on content type
    let body;
    const contentType = request.headers.get('content-type');

    if (contentType && contentType.includes('application/json')) {
      body = await request.json();
    } else {
      body = await request.text();
    }

    // Extract prompt from various formats
    const prompt = this.extractPrompt(body, request);

    // Client information
    const clientIP = this.extractClientIP(request);
    const userAgent = request.headers.get('user-agent') || 'unknown';

    // Request metadata
    const requestSize = this.calculateRequestSize(request, body);
    const timestamp = Date.now();

    return {
      method,
      endpoint,
      provider,
      prompt,
      body,
      clientIP,
      userAgent,
      requestSize,
      timestamp,
      headers: Object.fromEntries(request.headers.entries()),
      queryParams: Object.fromEntries(url.searchParams.entries())
    };
  }

  extractPrompt(body, request) {
    // Handle different input formats
    if (typeof body === 'string') {
      return body;
    }

    if (body && typeof body === 'object') {
      // Check common prompt fields
      return body.prompt || body.text || body.message || body.query || JSON.stringify(body);
    }

    // Fallback
    return '';
  }

  extractClientIP(request) {
    return request.headers.get('cf-connecting-ip') ||
           request.headers.get('x-forwarded-for') ||
           request.headers.get('x-real-ip') ||
           'unknown';
  }

  calculateRequestSize(request, body) {
    let size = 0;

    // Headers size
    for (const [key, value] of request.headers.entries()) {
      size += key.length + value.length;
    }

    // Body size
    if (typeof body === 'string') {
      size += body.length;
    } else if (body) {
      size += JSON.stringify(body).length;
    }

    return size;
  }

  async processRequestLogic(requestData, context) {
    // This would integrate with the provider manager and cache
    // For now, return a mock response
    return {
      result: `Processed request for provider: ${requestData.provider}`,
      provider: requestData.provider,
      cacheStatus: 'miss',
      tokens: requestData.prompt ? requestData.prompt.length : 0
    };
  }

  async handleError(error, request, responseTime) {
    // Enhanced error handling with recovery strategies
    const errorType = this.classifyError(error);
    const recoveryStrategy = this.determineRecoveryStrategy(errorType);

    // Log error details
    if (this.errorHandler) {
      await this.errorHandler.handle(error, this.extractClientIP(request), 'request-processing');
    }

    // Attempt recovery if appropriate
    if (recoveryStrategy === 'retry') {
      // Could implement retry logic here
    }

    // Build error response
    const errorResponse = {
      error: {
        type: errorType,
        message: this.getUserFriendlyMessage(error),
        code: error.status || 500,
        recovery: recoveryStrategy
      },
      timestamp: Date.now(),
      requestId: `req-${Date.now()}`
    };

    return new Response(JSON.stringify(errorResponse), {
      status: error.status || 500,
      headers: {
        'content-type': 'application/json',
        'x-error-type': errorType,
        'x-response-time': `${responseTime}ms`,
        'x-recovery-strategy': recoveryStrategy
      }
    });
  }

  classifyError(error) {
    if (error.message && error.message.includes('timeout')) return 'timeout';
    if (error.message && error.message.includes('rate limit')) return 'rate_limit';
    if (error.message && error.message.includes('auth')) return 'authentication';
    if (error.status === 429) return 'rate_limit';
    if (error.status >= 500) return 'server_error';
    if (error.status >= 400) return 'client_error';
    return 'unknown';
  }

  determineRecoveryStrategy(errorType) {
    switch (errorType) {
      case 'timeout': return 'retry';
      case 'rate_limit': return 'backoff';
      case 'server_error': return 'retry';
      case 'client_error': return 'fail';
      default: return 'fail';
    }
  }

  getUserFriendlyMessage(error) {
    const errorType = this.classifyError(error);
    switch (errorType) {
      case 'timeout': return 'Request timed out. Please try again.';
      case 'rate_limit': return 'Too many requests. Please wait before trying again.';
      case 'authentication': return 'Authentication failed. Please check your credentials.';
      case 'server_error': return 'Server error. Please try again later.';
      default: return 'An unexpected error occurred. Please try again.';
    }
  }

  updateAverageResponseTime(responseTime) {
    // Simple moving average calculation
    const alpha = 0.1; // Smoothing factor
    this.avgResponseTime = (alpha * responseTime) + ((1 - alpha) * this.avgResponseTime);
  }

  // Legacy method for backward compatibility
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