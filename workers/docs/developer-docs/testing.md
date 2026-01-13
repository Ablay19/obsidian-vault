# Testing Guide

## 🧪 Testing Enhanced Workers

This guide covers comprehensive testing strategies for the enhanced workers, including unit tests, integration tests, performance testing, and deployment verification.

## 🏗️ Test Environment Setup

### 1. Install Testing Dependencies

```bash
# Install testing framework
npm install --save-dev jest @types/jest
npm install --save-dev supertest
npm install --save-dev artillery # For load testing
npm install --save-dev clinic # For performance profiling
```

### 2. Configure Jest

Create `jest.config.js`:

```javascript
export default {
  testEnvironment: 'node',
  testMatch: ['**/__tests__/**/*.js', '**/?(*.)+(spec|test).js'],
  collectCoverageFrom: [
    'src/**/*.js',
    '!src/index.js', // Entry point
    '!**/node_modules/**'
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  },
  setupFilesAfterEnv: ['<rootDir>/test/setup.js'],
  testTimeout: 10000
};
```

### 3. Test Setup File

Create `test/setup.js`:

```javascript
// Global test setup
global.testConfig = {
  cacheSize: 10,
  rateLimitPerHour: 100,
  enableAnalytics: true,
  logLevel: 'error' // Reduce noise in tests
};

// Mock Cloudflare runtime
global.caches = {
  default: {
    put: jest.fn(),
    match: jest.fn(),
    delete: jest.fn()
  }
};

// Mock KV namespace
global.KV_NAMESPACE = {
  get: jest.fn(),
  put: jest.fn(),
  delete: jest.fn()
};

// Mock environment
global.mockEnv = {
  CACHE_KV: KV_NAMESPACE,
  ANALYTICS_KV: KV_NAMESPACE,
  RATE_LIMITER: {},
  ...testConfig
};

// Mock request/response
global.createMockRequest = (options = {}) => {
  return new Request('http://localhost/test', {
    method: options.method || 'GET',
    headers: options.headers || {},
    body: options.body
  });
};

global.createMockResponse = (options = {}) => {
  return new Response(options.body || '', {
    status: options.status || 200,
    headers: options.headers || {}
  });
};
```

## 🧩 Unit Testing

### 1. Analytics Testing

```javascript
// __tests__/analytics.test.js
import { WorkerAnalytics } from '../src/analytics.js';

describe('WorkerAnalytics', () => {
  let analytics;

  beforeEach(() => {
    analytics = new WorkerAnalytics({});
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('trackRequest', () => {
    test('should track successful requests', () => {
      analytics.trackRequest(25, true, '/api/chat');

      const report = analytics.getPerformanceReport();
      expect(report.totalRequests).toBe(1);
      expect(report.errorRate).toBe(0);
      expect(report.avgResponseTime).toBe(25);
    });

    test('should track failed requests', () => {
      analytics.trackRequest(45, false, '/api/chat');

      const report = analytics.getPerformanceReport();
      expect(report.totalRequests).toBe(1);
      expect(report.errorRate).toBe(100);
    });

    test('should maintain response time history', () => {
      // Track multiple requests
      analytics.trackRequest(10, true, '/api/chat');
      analytics.trackRequest(20, true, '/api/chat');
      analytics.trackRequest(30, true, '/api/chat');

      const report = analytics.getPerformanceReport();
      expect(report.totalRequests).toBe(3);
      expect(report.avgResponseTime).toBeCloseTo(20);
    });

    test('should limit response time history', () => {
      // Track more than 1000 requests
      for (let i = 0; i < 1100; i++) {
        analytics.trackRequest(i % 100, true, '/api/chat');
      }

      // Check that only recent 1000 are kept
      expect(analytics.metrics.responseTime.length).toBeLessThanOrEqual(1000);
    });
  });

  describe('calculateHealthScore', () => {
    test('should calculate perfect health score', () => {
      analytics.trackRequest(10, true, '/api/chat');

      const score = analytics.calculateHealthScore();
      expect(score).toBeGreaterThan(95);
    });

    test('should penalize high error rates', () => {
      analytics.trackRequest(50, false, '/api/chat');
      analytics.trackRequest(50, false, '/api/chat');

      const score = analytics.calculateHealthScore();
      expect(score).toBeLessThan(50);
    });
  });

  describe('getPerformanceReport', () => {
    test('should return comprehensive report', () => {
      analytics.trackRequest(25, true, '/api/chat');
      analytics.trackRequest(75, false, '/api/chat');

      const report = analytics.getPerformanceReport();

      expect(report).toHaveProperty('totalRequests', 2);
      expect(report).toHaveProperty('errorRate', 50);
      expect(report).toHaveProperty('avgResponseTime', 50);
      expect(report).toHaveProperty('healthScore');
      expect(report).toHaveProperty('cacheEfficiency', 0);
    });
  });
});
```

### 2. Cache Testing

```javascript
// __tests__/cache.test.js
import { SmartCache, CacheAnalytics } from '../src/cache.js';

describe('SmartCache', () => {
  let cache;
  let analytics;

  beforeEach(() => {
    analytics = new CacheAnalytics();
    cache = new SmartCache(10, analytics);
  });

  describe('basic operations', () => {
    test('should store and retrieve values', () => {
      cache.set('key1', 'value1');
      expect(cache.get('key1')).toBe('value1');
    });

    test('should return null for missing keys', () => {
      expect(cache.get('missing')).toBeNull();
    });

    test('should handle TTL expiration', async () => {
      cache.set('temp', 'value', 10); // 10ms TTL
      expect(cache.get('temp')).toBe('value');

      // Wait for expiration
      await new Promise(resolve => setTimeout(resolve, 15));
      expect(cache.get('temp')).toBeNull();
    });
  });

  describe('LRU eviction', () => {
    test('should evict least recently used items', () => {
      // Fill cache
      for (let i = 0; i < 12; i++) {
        cache.set(`key${i}`, `value${i}`);
      }

      // Cache should maintain max size
      expect(cache.cache.size).toBeLessThanOrEqual(10);
    });

    test('should prioritize recently accessed items', () => {
      cache.set('old', 'old_value');
      cache.set('new', 'new_value');

      // Access old item to make it recently used
      cache.get('old');

      // Fill cache to trigger eviction
      for (let i = 0; i < 10; i++) {
        cache.set(`filler${i}`, `filler_value${i}`);
      }

      // Old item should still be there
      expect(cache.get('old')).toBe('old_value');
      // New item might be evicted
      expect(cache.get('new')).toBeNull();
    });
  });

  describe('analytics integration', () => {
    test('should track cache hits', () => {
      cache.set('test', 'value');
      cache.get('test'); // Hit

      const stats = analytics.getReport();
      expect(stats.totalOperations).toBe(1);
      expect(stats.hitRate).toBe(1.0);
    });

    test('should track cache misses', () => {
      cache.get('missing'); // Miss

      const stats = analytics.getReport();
      expect(stats.totalOperations).toBe(1);
      expect(stats.hitRate).toBe(0);
    });
  });
});
```

### 3. Rate Limiter Testing

```javascript
// __tests__/rate-limiter.test.js
import { RateLimiter } from '../src/rate-limiter.js';

describe('RateLimiter', () => {
  let rateLimiter;

  beforeEach(() => {
    rateLimiter = new RateLimiter(global.KV_NAMESPACE);
  });

  describe('token bucket algorithm', () => {
    test('should allow requests within capacity', () => {
      const options = {
        algorithm: 'token-bucket',
        capacity: 10,
        refillRate: 1
      };

      // Should allow 10 requests immediately
      for (let i = 0; i < 10; i++) {
        expect(rateLimiter.allow('user1', options)).toBe(true);
      }

      // 11th request should be blocked
      expect(rateLimiter.allow('user1', options)).toBe(false);
    });

    test('should refill tokens over time', async () => {
      const options = {
        algorithm: 'token-bucket',
        capacity: 5,
        refillRate: 1
      };

      // Use all tokens
      for (let i = 0; i < 5; i++) {
        expect(rateLimiter.allow('user1', options)).toBe(true);
      }
      expect(rateLimiter.allow('user1', options)).toBe(false);

      // Wait for refill (simulate 2 seconds)
      await new Promise(resolve => setTimeout(resolve, 10));

      // Should allow one more request
      expect(rateLimiter.allow('user1', options)).toBe(true);
    });
  });

  describe('sliding window algorithm', () => {
    test('should limit requests in time window', () => {
      const options = {
        algorithm: 'sliding-window',
        capacity: 3,
        windowSize: 60
      };

      // Should allow 3 requests
      for (let i = 0; i < 3; i++) {
        expect(rateLimiter.allow('user1', options)).toBe(true);
      }

      // 4th request should be blocked
      expect(rateLimiter.allow('user1', options)).toBe(false);
    });
  });

  describe('graceful degradation', () => {
    test('should allow some requests when rate limited', () => {
      const options = {
        algorithm: 'token-bucket',
        capacity: 1,
        refillRate: 0.1
      };

      // Use up capacity
      expect(rateLimiter.allow('user1', options)).toBe(true);
      expect(rateLimiter.allow('user1', options)).toBe(false);

      // Test graceful degradation
      let allowed = 0;
      let blocked = 0;

      for (let i = 0; i < 100; i++) {
        if (rateLimiter.allowWithDegradation('user1', options)) {
          allowed++;
        } else {
          blocked++;
        }
      }

      // Should allow about 10% even when rate limited
      expect(allowed).toBeGreaterThan(5);
      expect(blocked).toBeGreaterThan(allowed);
    });
  });

  describe('statistics', () => {
    test('should track request statistics', () => {
      const options = { algorithm: 'token-bucket', capacity: 1 };

      rateLimiter.allow('user1', options); // Allowed
      rateLimiter.allow('user1', options); // Blocked

      const stats = rateLimiter.getStats();
      expect(stats.totalRequests).toBe(2);
      expect(stats.blockedRequests).toBe(1);
    });
  });
});
```

## 🔗 Integration Testing

### 1. End-to-End Request Testing

```javascript
// __tests__/integration/e2e.test.js
import { createMockRequest, createMockResponse, mockEnv } from '../../test/setup.js';
import worker from '../../src/index.js';

describe('End-to-End Integration', () => {
  test('should handle complete AI request flow', async () => {
    const request = createMockRequest({
      method: 'POST',
      url: 'https://worker.example.com/',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer test-token'
      },
      body: JSON.stringify({
        prompt: "Hello, how are you?",
        model: "gpt-3.5-turbo",
        maxTokens: 50
      })
    });

    const response = await worker.fetch(request, mockEnv, {});

    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data).toHaveProperty('result');
    expect(data).toHaveProperty('provider');
    expect(data).toHaveProperty('tokens');
    expect(data).toHaveProperty('processingTime');
  });

  test('should handle rate limiting', async () => {
    const requests = [];

    // Create many requests quickly
    for (let i = 0; i < 150; i++) {
      const request = createMockRequest({
        method: 'POST',
        headers: { 'CF-Connecting-IP': '192.168.1.1' }
      });
      requests.push(worker.fetch(request, mockEnv, {}));
    }

    const responses = await Promise.all(requests);
    const rateLimited = responses.filter(r => r.status === 429);

    expect(rateLimited.length).toBeGreaterThan(0);
  });

  test('should serve cached responses', async () => {
    const request1 = createMockRequest({
      method: 'POST',
      body: JSON.stringify({ prompt: "Test prompt", cache: true })
    });

    const request2 = createMockRequest({
      method: 'POST',
      body: JSON.stringify({ prompt: "Test prompt", cache: true })
    });

    // First request
    const response1 = await worker.fetch(request1, mockEnv, {});
    expect(response1.status).toBe(200);

    // Second request should be cached
    const response2 = await worker.fetch(request2, mockEnv, {});
    expect(response2.status).toBe(200);

    const data2 = await response2.json();
    expect(data2.cache).toHaveProperty('hit', true);
  });
});
```

### 2. Component Integration Testing

```javascript
// __tests__/integration/components.test.js
import { WorkerAnalytics } from '../../src/analytics.js';
import { SmartCache } from '../../src/cache.js';
import { RateLimiter } from '../../src/rate-limiter.js';
import { ProviderManager } from '../../src/provider-manager.js';

describe('Component Integration', () => {
  let analytics;
  let cache;
  let rateLimiter;
  let providerManager;

  beforeEach(() => {
    analytics = new WorkerAnalytics({});
    cache = new SmartCache(10);
    rateLimiter = new RateLimiter(global.KV_NAMESPACE);
    providerManager = new ProviderManager({}, {}, analytics);
  });

  test('should integrate analytics with cache', () => {
    // Set up cache with analytics
    const cacheWithAnalytics = new SmartCache(10, {
      recordCacheHit: jest.fn(),
      recordCacheMiss: jest.fn()
    });

    cacheWithAnalytics.set('test', 'value');
    cacheWithAnalytics.get('test'); // Hit
    cacheWithAnalytics.get('missing'); // Miss

    // Analytics should have been called
    expect(cacheWithAnalytics.analytics.recordCacheHit).toHaveBeenCalled();
    expect(cacheWithAnalytics.analytics.recordCacheMiss).toHaveBeenCalled();
  });

  test('should integrate rate limiting with analytics', () => {
    const options = { algorithm: 'token-bucket', capacity: 1 };

    // First request allowed
    expect(rateLimiter.allow('user1', options)).toBe(true);

    // Second request blocked
    expect(rateLimiter.allow('user1', options)).toBe(false);

    // Check analytics
    const stats = rateLimiter.getStats();
    expect(stats.totalRequests).toBe(2);
    expect(stats.blockedRequests).toBe(1);
  });

  test('should integrate all components in request flow', async () => {
    // Mock complete request flow
    const mockRequest = {
      prompt: "Test request",
      clientIP: "127.0.0.1"
    };

    // Rate limiting check
    expect(rateLimiter.allow('test-user')).toBe(true);

    // Cache check (miss)
    expect(cache.get('test-request')).toBeNull();

    // Provider selection
    const provider = await providerManager.selectProviderWithLoadBalancing(mockRequest);
    expect(provider).toBeDefined();

    // Analytics tracking
    analytics.trackRequest(25, true, '/api/chat');

    const report = analytics.getPerformanceReport();
    expect(report.totalRequests).toBe(1);
    expect(report.errorRate).toBe(0);
  });
});
```

## 🚀 Performance Testing

### 1. Load Testing with Artillery

Create `test/load/load-test.yml`:

```yaml
config:
  target: 'https://your-worker.workers.dev'
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 120
      arrivalRate: 50
      name: "Load testing"
    - duration: 60
      arrivalRate: 100
      name: "Stress testing"

scenarios:
  - name: "AI Request"
    weight: 70
    request:
      method: POST
      url: "/"
      headers:
        Content-Type: application/json
      json:
        prompt: "Write a short story about AI"
        maxTokens: 100

  - name: "Cache Hit Test"
    weight: 20
    request:
      method: POST
      url: "/"
      headers:
        Content-Type: application/json
      json:
        prompt: "Hello world"
        maxTokens: 10

  - name: "Health Check"
    weight: 10
    request:
      url: "/health"
      method: GET
```

Run load tests:

```bash
# Install artillery
npm install -g artillery

# Run load test
artillery run test/load/load-test.yml

# Generate report
artillery report report.json
```

### 2. Performance Profiling

```javascript
// test/performance/profile.test.js
import { WorkerProfiler } from '../../src/profiler.js';

describe('Performance Profiling', () => {
  let profiler;

  beforeEach(() => {
    profiler = new WorkerProfiler();
  });

  test('should profile function execution', () => {
    const profileId = profiler.startProfile('test-function');

    // Simulate work
    let result = 0;
    for (let i = 0; i < 10000; i++) {
      result += Math.sin(i);
    }

    const profileResult = profiler.endProfile(profileId);

    expect(profileResult).toHaveProperty('name', 'test-function');
    expect(profileResult).toHaveProperty('duration');
    expect(profileResult).toHaveProperty('memoryDelta');
    expect(profileResult.duration).toBeGreaterThan(0);
  });

  test('should handle checkpoints', () => {
    const profileId = profiler.startProfile('checkpoint-test');

    // Phase 1
    profiler.addCheckpoint('phase1-start');
    for (let i = 0; i < 1000; i++) {
      Math.sqrt(i);
    }
    profiler.addCheckpoint('phase1-end');

    // Phase 2
    profiler.addCheckpoint('phase2-start');
    for (let i = 0; i < 2000; i++) {
      Math.pow(i, 2);
    }
    profiler.addCheckpoint('phase2-end');

    const profileResult = profiler.endProfile(profileId);

    expect(profileResult.checkpoints).toHaveLength(4);
    expect(profileResult.checkpoints[0]).toHaveProperty('label', 'phase1-start');
  });

  test('should export profiling data', () => {
    profiler.startProfile('export-test');
    profiler.endProfile();

    const exportData = profiler.exportProfiles('json');
    const parsed = JSON.parse(exportData);

    expect(parsed).toHaveProperty('stats');
    expect(parsed).toHaveProperty('completedProfiles');
  });
});
```

### 3. Memory Leak Testing

```javascript
// test/performance/memory-leak.test.js
describe('Memory Leak Detection', () => {
  test('should not leak memory during repeated requests', async () => {
    const initialMemory = process.memoryUsage().heapUsed;

    // Simulate many requests
    for (let i = 0; i < 1000; i++) {
      const request = createMockRequest({
        method: 'POST',
        body: JSON.stringify({ prompt: `Request ${i}`, maxTokens: 10 })
      });

      await worker.fetch(request, mockEnv, {});
    }

    const finalMemory = process.memoryUsage().heapUsed;
    const memoryIncrease = finalMemory - initialMemory;

    // Allow some memory increase but not excessive
    expect(memoryIncrease).toBeLessThan(10 * 1024 * 1024); // 10MB limit
  });

  test('should cleanup expired cache entries', async () => {
    const cache = new SmartCache(100);

    // Add many entries with short TTL
    for (let i = 0; i < 50; i++) {
      cache.set(`key${i}`, `value${i}`, 10); // 10ms TTL
    }

    // Wait for expiration
    await new Promise(resolve => setTimeout(resolve, 20));

    // Trigger cleanup
    cache.cleanup();

    // Should have cleaned up expired entries
    expect(cache.cache.size).toBeLessThan(50);
  });
});
```

## 🔒 Security Testing

### 1. Input Validation Testing

```javascript
// __tests__/security/input-validation.test.js
describe('Input Validation Security', () => {
  test('should reject oversized prompts', async () => {
    const largePrompt = 'x'.repeat(100000); // 100KB
    const request = createMockRequest({
      method: 'POST',
      body: JSON.stringify({ prompt: largePrompt })
    });

    const response = await worker.fetch(request, mockEnv, {});
    expect(response.status).toBe(400);
  });

  test('should sanitize HTML in prompts', async () => {
    const maliciousPrompt = '<script>alert("xss")</script>Hello';
    const request = createMockRequest({
      method: 'POST',
      body: JSON.stringify({ prompt: maliciousPrompt })
    });

    const response = await worker.fetch(request, mockEnv, {});
    expect(response.status).toBe(200);

    const data = await response.json();
    expect(data.result).not.toContain('<script>');
  });

  test('should validate API keys', async () => {
    const request = createMockRequest({
      method: 'POST',
      headers: { 'Authorization': 'Bearer invalid-key' }
    });

    const response = await worker.fetch(request, mockEnv, {});
    expect(response.status).toBe(401);
  });
});
```

### 2. Rate Limiting Security Testing

```javascript
// __tests__/security/rate-limiting.test.js
describe('Rate Limiting Security', () => {
  test('should prevent brute force attacks', async () => {
    const requests = [];

    // Simulate brute force attack
    for (let i = 0; i < 1000; i++) {
      const request = createMockRequest({
        method: 'POST',
        headers: { 'CF-Connecting-IP': '192.168.1.100' }
      });
      requests.push(worker.fetch(request, mockEnv, {}));
    }

    const responses = await Promise.all(requests);
    const blockedResponses = responses.filter(r => r.status === 429);

    expect(blockedResponses.length).toBeGreaterThan(900); // >90% blocked
  });

  test('should handle distributed attacks', async () => {
    const ips = ['192.168.1.1', '192.168.1.2', '192.168.1.3'];
    const requests = [];

    // Simulate distributed attack
    for (let i = 0; i < 300; i++) {
      const ip = ips[i % ips.length];
      const request = createMockRequest({
        method: 'POST',
        headers: { 'CF-Connecting-IP': ip }
      });
      requests.push(worker.fetch(request, mockEnv, {}));
    }

    const responses = await Promise.all(requests);
    const successfulResponses = responses.filter(r => r.status === 200);

    // Should allow reasonable requests from different IPs
    expect(successfulResponses.length).toBeGreaterThan(100);
  });
});
```

## 🚀 CI/CD Testing

### 1. GitHub Actions Configuration

Create `.github/workflows/test.yml`:

```yaml
name: Test Enhanced Workers

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'

    - name: Install dependencies
      run: npm ci

    - name: Run unit tests
      run: npm run test:unit

    - name: Run integration tests
      run: npm run test:integration

    - name: Run security tests
      run: npm run test:security

    - name: Generate coverage report
      run: npm run test:coverage

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage/lcov.info

  performance:
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'

    steps:
    - uses: actions/checkout@v3

    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'

    - name: Install dependencies
      run: npm ci

    - name: Run performance tests
      run: npm run test:performance

    - name: Comment performance results
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');
          const results = JSON.parse(fs.readFileSync('./performance-results.json', 'utf8'));

          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: `## 🚀 Performance Test Results

**Response Time:** ${results.avgResponseTime}ms (target: <50ms)
**Cache Hit Rate:** ${(results.cacheHitRate * 100).toFixed(1)}% (target: >85%)
**Memory Usage:** ${results.memoryUsage}MB (target: <10MB)
**Error Rate:** ${(results.errorRate * 100).toFixed(2)}% (target: <1%)

${results.passed ? '✅ All targets met!' : '⚠️ Some targets not met'}`

  deploy-staging:
    needs: [test, performance]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'

    steps:
    - uses: actions/checkout@v3

    - name: Deploy to staging
      run: npx wrangler deploy --env staging
```

### 2. Test Script Configuration

Update `package.json`:

```json
{
  "scripts": {
    "test": "jest",
    "test:unit": "jest --testPathPattern=unit",
    "test:integration": "jest --testPathPattern=integration",
    "test:security": "jest --testPathPattern=security",
    "test:performance": "jest --testPathPattern=performance",
    "test:coverage": "jest --coverage",
    "test:watch": "jest --watch",
    "test:e2e": "jest --testPathPattern=e2e",
    "test:load": "artillery run test/load/load-test.yml",
    "lint": "eslint src/**/*.js",
    "lint:fix": "eslint src/**/*.js --fix",
    "type-check": "tsc --noEmit"
  }
}
```

## 📊 Test Reporting

### 1. Coverage Reports

Generate detailed coverage reports:

```bash
npm run test:coverage
```

### 2. Performance Benchmarks

```javascript
// benchmarks/response-time.bench.js
import { Bench } from 'tinybench';

const bench = new Bench();

bench
  .add('Simple AI Request', async () => {
    const request = createMockRequest({
      method: 'POST',
      body: JSON.stringify({ prompt: "Hello", maxTokens: 10 })
    });
    await worker.fetch(request, mockEnv, {});
  })
  .add('Cached AI Request', async () => {
    const request = createMockRequest({
      method: 'POST',
      body: JSON.stringify({ prompt: "Hello", cache: true })
    });
    await worker.fetch(request, mockEnv, {});
  });

await bench.run();

console.table(bench.table());
```

### 3. Test Results Dashboard

Create a test results dashboard:

```javascript
// test/dashboard/test-dashboard.js
import fs from 'fs';

export class TestDashboard {
  constructor() {
    this.results = {
      unit: { passed: 0, failed: 0, coverage: 0 },
      integration: { passed: 0, failed: 0 },
      performance: { passed: 0, failed: 0 },
      security: { passed: 0, failed: 0 }
    };
  }

  updateResults(type, results) {
    this.results[type] = {
      passed: results.numPassedTests || 0,
      failed: results.numFailedTests || 0,
      coverage: results.coverage || 0
    };
  }

  generateReport() {
    const totalPassed = Object.values(this.results).reduce((sum, r) => sum + r.passed, 0);
    const totalFailed = Object.values(this.results).reduce((sum, r) => sum + r.failed, 0);

    return {
      summary: {
        totalTests: totalPassed + totalFailed,
        passed: totalPassed,
        failed: totalFailed,
        passRate: totalPassed / (totalPassed + totalFailed)
      },
      details: this.results,
      timestamp: new Date().toISOString()
    };
  }

  saveReport() {
    const report = this.generateReport();
    fs.writeFileSync('./test-results.json', JSON.stringify(report, null, 2));
  }
}
```

## 🎯 Testing Best Practices

### Test Organization
- Keep tests close to the code they test
- Use descriptive test names
- Group related tests in describe blocks
- Use beforeEach/afterEach for setup/cleanup

### Mocking Strategy
- Mock external dependencies (APIs, databases)
- Use realistic test data
- Avoid over-mocking
- Test error conditions

### Performance Testing
- Run performance tests separately from unit tests
- Use realistic data volumes
- Measure memory usage
- Test under load conditions

### CI/CD Integration
- Run tests on every PR
- Fail builds on test failures
- Monitor test performance over time
- Alert on significant regressions

This comprehensive testing strategy ensures the enhanced workers are reliable, secure, and performant across all use cases.