# Performance Workshop

## ⚡ Performance Workshop Overview

This intensive hands-on workshop focuses on advanced performance optimization techniques for enhanced Cloudflare Workers. Participants will learn to identify bottlenecks, implement optimizations, and measure performance improvements through practical exercises.

## 🎯 Workshop Objectives

### Primary Goals
- Master performance profiling techniques
- Identify and resolve performance bottlenecks
- Implement advanced optimization strategies
- Measure and quantify performance improvements
- Develop systematic performance optimization workflows

### Expected Outcomes
- 50-80% performance improvement in response times
- 20-40% reduction in resource usage
- Comprehensive performance monitoring setup
- Optimized caching and rate limiting configurations
- Performance benchmarking and regression testing

## 📚 Workshop Agenda

### Day 1: Performance Foundations

#### Morning: Performance Analysis Tools

```yaml
duration: 3 hours
format: Interactive Lab

objectives:
  - Set up performance profiling tools
  - Learn performance measurement techniques
  - Understand performance metrics interpretation
  - Practice bottleneck identification

hands-on_exercises:
  1. Tool Setup and Configuration
  2. Baseline Performance Measurement
  3. Profiling a Slow API Endpoint
  4. Memory Leak Detection
```

#### Exercise 1.1: Tool Setup

```bash
# Install performance monitoring tools
npm install -D lighthouse puppeteer artillery autocannon

# Set up performance monitoring
mkdir performance-tools
cd performance-tools

# Create performance measurement script
cat > measure-performance.js << 'EOF'
const autocannon = require('autocannon');

async function measurePerformance(url, options = {}) {
  const instance = autocannon({
    url,
    connections: options.connections || 10,
    duration: options.duration || 10,
    headers: options.headers || {},
    ...options
  });

  autocannon.track(instance);

  return new Promise((resolve, reject) => {
    instance.on('done', (results) => {
      resolve({
        requests: results.requests,
        throughput: results.throughput,
        latency: results.latency,
        errors: results.errors
      });
    });

    instance.on('error', reject);
  });
}

// Export for use in other scripts
module.exports = { measurePerformance };
EOF
```

#### Exercise 1.2: Baseline Measurement

```javascript
// performance-baseline.js
const { measurePerformance } = require('./measure-performance');

async function runBaselineTests() {
  console.log('Running baseline performance tests...\n');

  const tests = [
    { name: 'Health Check', url: 'https://api.yourdomain.com/health' },
    { name: 'Chat API', url: 'https://api.yourdomain.com/chat', method: 'POST' },
    { name: 'Analytics API', url: 'https://api.yourdomain.com/analytics' }
  ];

  for (const test of tests) {
    console.log(`Testing: ${test.name}`);
    try {
      const results = await measurePerformance(test.url, {
        connections: 10,
        duration: 30,
        method: test.method || 'GET'
      });

      console.log(`  Requests/sec: ${results.throughput.average}`);
      console.log(`  Latency (mean): ${results.latency.average}ms`);
      console.log(`  Latency (p95): ${results.latency.p95}ms`);
      console.log(`  Errors: ${results.errors}\n`);
    } catch (error) {
      console.error(`  Error: ${error.message}\n`);
    }
  }
}

runBaselineTests().catch(console.error);
```

#### Afternoon: Caching Optimization

```yaml
duration: 3 hours
format: Workshop

objectives:
  - Understand caching strategies
  - Implement advanced caching techniques
  - Optimize cache key generation
  - Measure cache effectiveness

hands-on_exercises:
  1. Cache Analysis and Optimization
  2. Cache Key Design Workshop
  3. Cache Preloading Implementation
  4. Cache Performance Measurement
```

#### Exercise 1.3: Cache Analysis

```javascript
// cache-analysis.js
class CacheAnalyzer {
  constructor(cache) {
    this.cache = cache;
    this.metrics = {
      hits: 0,
      misses: 0,
      evictions: 0,
      sizeHistory: []
    };
  }

  analyze() {
    const metrics = this.cache.getMetrics();
    const analysis = {
      hitRate: metrics.hitRate,
      size: metrics.size,
      evictions: this.metrics.evictions,
      recommendations: this.generateRecommendations(metrics)
    };

    return analysis;
  }

  generateRecommendations(metrics) {
    const recommendations = [];

    if (metrics.hitRate < 0.7) {
      recommendations.push({
        priority: 'high',
        issue: 'Low cache hit rate',
        solution: 'Increase cache TTL or optimize cache keys',
        impact: 'Can improve hit rate by 20-30%'
      });
    }

    if (metrics.size > 1000) {
      recommendations.push({
        priority: 'medium',
        issue: 'Large cache size',
        solution: 'Implement LRU eviction or reduce TTL',
        impact: 'Can reduce memory usage by 20-50%'
      });
    }

    return recommendations;
  }

  monitor(interval = 60000) {
    setInterval(() => {
      this.metrics.sizeHistory.push({
        timestamp: Date.now(),
        size: this.cache.size
      });

      // Keep last 24 hours of data
      const oneDay = 24 * 60 * 60 * 1000;
      this.metrics.sizeHistory = this.metrics.sizeHistory
        .filter(entry => Date.now() - entry.timestamp < oneDay);
    }, interval);
  }
}

// Usage example
const cacheAnalyzer = new CacheAnalyzer(cache);
cacheAnalyzer.monitor();

setInterval(() => {
  const analysis = cacheAnalyzer.analyze();
  console.log('Cache Analysis:', analysis);
}, 300000); // Every 5 minutes
```

### Day 2: Advanced Optimization

#### Morning: Rate Limiting Optimization

```yaml
duration: 3 hours
format: Interactive Lab

objectives:
  - Understand rate limiting algorithms
  - Implement adaptive rate limiting
  - Optimize rate limiting performance
  - Measure rate limiting effectiveness

hands-on_exercises:
  1. Rate Limiting Analysis
  2. Algorithm Comparison
  3. Adaptive Rate Limiting Implementation
  4. Rate Limiting Performance Testing
```

#### Exercise 2.1: Rate Limiting Optimization

```javascript
// adaptive-rate-limiter.js
export class AdaptiveRateLimiter {
  constructor(config = {}) {
    this.config = {
      baseLimit: config.baseLimit || 100,
      burstLimit: config.burstLimit || 20,
      windowSize: config.windowSize || 60000,
      adaptationInterval: config.adaptationInterval || 300000, // 5 minutes
      ...config
    };

    this.buckets = new Map();
    this.metrics = {
      totalRequests: 0,
      deniedRequests: 0,
      adaptationHistory: []
    };

    // Start adaptive adjustment
    this.startAdaptation();
  }

  async checkLimit(userId) {
    this.metrics.totalRequests++;
    const now = Date.now();
    const bucket = this.getOrCreateBucket(userId, now);

    // Check burst allowance
    if (bucket.burstTokens > 0) {
      bucket.burstTokens--;
      return { allowed: true, remaining: bucket.burstTokens };
    }

    // Check window limit
    if (bucket.requests < bucket.limit) {
      bucket.requests++;
      return { allowed: true, remaining: bucket.limit - bucket.requests };
    }

    // Request denied
    this.metrics.deniedRequests++;
    return {
      allowed: false,
      remaining: 0,
      resetTime: new Date(bucket.windowStart + this.config.windowSize).toISOString()
    };
  }

  getOrCreateBucket(userId, now) {
    let bucket = this.buckets.get(userId);

    if (!bucket || now - bucket.windowStart > this.config.windowSize) {
      bucket = {
        requests: 0,
        windowStart: now,
        burstTokens: this.config.burstLimit,
        limit: this.calculateDynamicLimit(userId)
      };
      this.buckets.set(userId, bucket);
    }

    return bucket;
  }

  calculateDynamicLimit(userId) {
    // Base limit with user-specific adjustments
    let limit = this.config.baseLimit;

    // Premium users get higher limits
    if (this.isPremiumUser(userId)) {
      limit *= 2;
    }

    // Reduce limit during high load
    const loadFactor = this.getSystemLoadFactor();
    if (loadFactor > 0.8) {
      limit *= 0.8;
    }

    // Increase limit during low load
    if (loadFactor < 0.3) {
      limit *= 1.2;
    }

    return Math.floor(limit);
  }

  isPremiumUser(userId) {
    // Check if user has premium subscription
    return userId.startsWith('premium-');
  }

  getSystemLoadFactor() {
    // Calculate system load based on recent metrics
    const recentMetrics = this.metrics.adaptationHistory.slice(-10);
    if (recentMetrics.length === 0) return 0.5;

    const avgDenialRate = recentMetrics.reduce(
      (sum, m) => sum + (m.denied / m.total), 0
    ) / recentMetrics.length;

    return Math.min(avgDenialRate * 2, 1); // Scale to 0-1
  }

  startAdaptation() {
    setInterval(() => {
      this.adaptLimits();
    }, this.config.adaptationInterval);
  }

  adaptLimits() {
    const currentMetrics = {
      timestamp: Date.now(),
      total: this.metrics.totalRequests,
      denied: this.metrics.deniedRequests,
      denialRate: this.metrics.deniedRequests / this.metrics.totalRequests
    };

    this.metrics.adaptationHistory.push(currentMetrics);

    // Keep last 100 adaptation points
    if (this.metrics.adaptationHistory.length > 100) {
      this.metrics.adaptationHistory.shift();
    }

    // Adjust base limits based on patterns
    this.adjustBaseLimits(currentMetrics);
  }

  adjustBaseLimits(metrics) {
    const denialRate = metrics.denialRate;

    // If denial rate is too high, increase limits
    if (denialRate > 0.1) {
      this.config.baseLimit = Math.min(this.config.baseLimit * 1.1, 1000);
      console.log(`Increased base limit to ${this.config.baseLimit} due to high denial rate`);
    }

    // If denial rate is very low, decrease limits to save resources
    if (denialRate < 0.01) {
      this.config.baseLimit = Math.max(this.config.baseLimit * 0.95, 10);
      console.log(`Decreased base limit to ${this.config.baseLimit} due to low denial rate`);
    }
  }

  getMetrics() {
    const denialRate = this.metrics.totalRequests > 0
      ? this.metrics.deniedRequests / this.metrics.totalRequests
      : 0;

    return {
      totalRequests: this.metrics.totalRequests,
      deniedRequests: this.metrics.deniedRequests,
      denialRate,
      activeBuckets: this.buckets.size,
      currentBaseLimit: this.config.baseLimit,
      adaptationHistory: this.metrics.adaptationHistory.slice(-10)
    };
  }
}
```

#### Afternoon: Memory and CPU Optimization

```yaml
duration: 3 hours
format: Interactive Lab

objectives:
  - Optimize memory usage
  - Reduce CPU overhead
  - Implement efficient data structures
  - Monitor resource usage

hands-on_exercises:
  1. Memory Profiling and Optimization
  2. CPU Usage Analysis
  3. Data Structure Optimization
  4. Resource Monitoring Implementation
```

#### Exercise 2.2: Memory Optimization

```javascript
// memory-optimizer.js
export class MemoryOptimizer {
  constructor() {
    this.snapshots = [];
    this.optimizationStrategies = {
      objectPooling: this.optimizeObjectPooling.bind(this),
      stringInterning: this.optimizeStringInterning.bind(this),
      garbageCollection: this.optimizeGarbageCollection.bind(this),
      dataStructureChoice: this.optimizeDataStructures.bind(this)
    };
  }

  takeMemorySnapshot(label) {
    const snapshot = {
      label,
      timestamp: Date.now(),
      memory: {
        used: performance.memory?.usedJSHeapSize || 0,
        total: performance.memory?.totalJSHeapSize || 0,
        limit: performance.memory?.jsHeapSizeLimit || 0
      },
      objects: this.countObjects()
    };

    this.snapshots.push(snapshot);
    return snapshot;
  }

  countObjects() {
    // Count different types of objects
    const objectCounts = {
      arrays: 0,
      objects: 0,
      functions: 0,
      strings: 0
    };

    // This is a simplified implementation
    // In a real scenario, you'd use more sophisticated object counting
    return objectCounts;
  }

  optimizeObjectPooling() {
    // Implement object pooling for frequently created objects
    class ObjectPool {
      constructor(factory, reset, initialSize = 10) {
        this.factory = factory;
        this.reset = reset;
        this.pool = [];
        this.active = new Set();

        // Pre-populate pool
        for (let i = 0; i < initialSize; i++) {
          this.pool.push(factory());
        }
      }

      acquire() {
        let obj = this.pool.pop();
        if (!obj) {
          obj = this.factory();
        }
        this.active.add(obj);
        return obj;
      }

      release(obj) {
        if (this.active.has(obj)) {
          this.reset(obj);
          this.active.delete(obj);
          this.pool.push(obj);
        }
      }

      getStats() {
        return {
          poolSize: this.pool.length,
          activeCount: this.active.size,
          totalCreated: this.pool.length + this.active.size
        };
      }
    }

    return ObjectPool;
  }

  optimizeStringInterning() {
    // Implement string interning for repeated strings
    class StringInterner {
      constructor() {
        this.strings = new Map();
        this.stats = { hits: 0, misses: 0 };
      }

      intern(str) {
        if (this.strings.has(str)) {
          this.stats.hits++;
          return this.strings.get(str);
        }

        this.stats.misses++;
        this.strings.set(str, str);
        return str;
      }

      getStats() {
        const total = this.stats.hits + this.stats.misses;
        return {
          ...this.stats,
          hitRate: total > 0 ? this.stats.hits / total : 0,
          uniqueStrings: this.strings.size
        };
      }
    }

    return StringInterner;
  }

  optimizeGarbageCollection() {
    // Implement strategies to reduce GC pressure
    class GCManager {
      constructor() {
        this.largeObjects = new WeakMap();
        this.cleanupTasks = [];
      }

      registerLargeObject(obj, cleanup) {
        this.largeObjects.set(obj, cleanup);
      }

      suggestGCOptimization() {
        const recommendations = [];

        // Check for large object usage
        if (this.largeObjects.size > 10) {
          recommendations.push({
            type: 'large_objects',
            message: 'Consider using streaming for large data processing',
            impact: 'High'
          });
        }

        // Check for cleanup task backlog
        if (this.cleanupTasks.length > 100) {
          recommendations.push({
            type: 'cleanup_backlog',
            message: 'Execute pending cleanup tasks',
            impact: 'Medium'
          });
        }

        return recommendations;
      }

      forceGC() {
        // Force garbage collection if available
        if (global.gc) {
          global.gc();
          console.log('Manual GC executed');
        }
      }
    }

    return GCManager;
  }

  optimizeDataStructures() {
    // Choose optimal data structures based on usage patterns
    class DataStructureOptimizer {
      analyzeUsage(pattern) {
        const recommendations = [];

        switch (pattern.type) {
          case 'frequent_lookups':
            if (pattern.size > 1000) {
              recommendations.push({
                structure: 'Map',
                reason: 'Better performance for large datasets',
                performance_gain: '20-50%'
              });
            }
            break;

          case 'ordered_operations':
            recommendations.push({
              structure: 'SortedMap or SkipList',
              reason: 'Maintains order for range queries',
              performance_gain: '30-70%'
            });
            break;

          case 'set_operations':
            recommendations.push({
              structure: 'Set or BloomFilter',
              reason: 'Optimized for membership testing',
              performance_gain: '10-40%'
            });
            break;
        }

        return recommendations;
      }
    }

    return DataStructureOptimizer;
  }

  runOptimizationSuite() {
    const results = {};

    for (const [name, optimizer] of Object.entries(this.optimizationStrategies)) {
      try {
        const beforeSnapshot = this.takeMemorySnapshot(`${name}_before`);
        const optimizationResult = optimizer();
        const afterSnapshot = this.takeMemorySnapshot(`${name}_after`);

        results[name] = {
          before: beforeSnapshot,
          after: afterSnapshot,
          memorySavings: beforeSnapshot.memory.used - afterSnapshot.memory.used,
          result: optimizationResult
        };
      } catch (error) {
        results[name] = { error: error.message };
      }
    }

    return results;
  }

  generateOptimizationReport() {
    const results = this.runOptimizationSuite();
    const report = {
      timestamp: new Date().toISOString(),
      optimizations: {},
      summary: {
        totalMemorySaved: 0,
        optimizationsApplied: 0,
        recommendations: []
      }
    };

    for (const [name, result] of Object.entries(results)) {
      if (result.error) {
        report.optimizations[name] = { status: 'failed', error: result.error };
      } else {
        report.optimizations[name] = {
          status: 'applied',
          memorySavings: result.memorySavings,
          beforeMemory: result.before.memory.used,
          afterMemory: result.after.memory.used
        };

        report.summary.totalMemorySaved += result.memorySavings;
        report.summary.optimizationsApplied++;
      }
    }

    return report;
  }
}
```

### Day 3: Performance Testing and Monitoring

#### Morning: Load Testing

```yaml
duration: 3 hours
format: Interactive Lab

objectives:
  - Design comprehensive load tests
  - Execute load testing scenarios
  - Analyze load test results
  - Identify breaking points

hands-on_exercises:
  1. Load Test Design and Execution
  2. Stress Testing Scenarios
  3. Performance Regression Testing
  4. Scalability Testing
```

#### Exercise 3.1: Advanced Load Testing

```javascript
// advanced-load-test.js
const autocannon = require('autocannon');

class AdvancedLoadTester {
  constructor(config = {}) {
    this.config = {
      baseUrl: config.baseUrl || 'https://api.yourdomain.com',
      scenarios: config.scenarios || [],
      monitoring: config.monitoring || {},
      ...config
    };

    this.results = [];
    this.monitoring = new PerformanceMonitor(this.config.monitoring);
  }

  async runComprehensiveTest() {
    console.log('Starting comprehensive load test suite...\n');

    const testResults = [];

    // Run baseline test
    console.log('Running baseline test...');
    const baseline = await this.runBaselineTest();
    testResults.push({ name: 'baseline', ...baseline });

    // Run scenario tests
    for (const scenario of this.config.scenarios) {
      console.log(`Running scenario: ${scenario.name}`);
      const result = await this.runScenarioTest(scenario);
      testResults.push(result);
    }

    // Run stress test
    console.log('Running stress test...');
    const stress = await this.runStressTest();
    testResults.push({ name: 'stress', ...stress });

    // Run endurance test
    console.log('Running endurance test...');
    const endurance = await this.runEnduranceTest();
    testResults.push({ name: 'endurance', ...endurance });

    // Generate comprehensive report
    const report = this.generateComprehensiveReport(testResults);
    this.saveReport(report);

    return report;
  }

  async runBaselineTest() {
    return await this.runLoadTest({
      name: 'baseline',
      connections: 10,
      duration: 60,
      url: `${this.config.baseUrl}/health`
    });
  }

  async runScenarioTest(scenario) {
    const results = [];

    for (const phase of scenario.phases) {
      const phaseResult = await this.runLoadTest({
        name: `${scenario.name}_${phase.name}`,
        connections: phase.connections,
        duration: phase.duration,
        url: phase.url || scenario.url,
        method: phase.method || 'GET',
        body: phase.body,
        headers: phase.headers
      });

      results.push(phaseResult);

      // Brief pause between phases
      if (phase.pause) {
        await this.sleep(phase.pause);
      }
    }

    return {
      name: scenario.name,
      type: 'scenario',
      phases: results,
      summary: this.summarizeScenarioResults(results)
    };
  }

  async runStressTest() {
    const stressLevels = [50, 100, 200, 500, 1000];

    for (const connections of stressLevels) {
      console.log(`Testing ${connections} connections...`);

      try {
        const result = await this.runLoadTest({
          name: `stress_${connections}`,
          connections,
          duration: 30,
          url: `${this.config.baseUrl}/chat`
        });

        // Check for breaking points
        if (result.errors > result.requests.total * 0.1) {
          console.log(`Breaking point detected at ${connections} connections`);
          return {
            breakingPoint: connections,
            result,
            status: 'failed'
          };
        }

        await this.sleep(5000); // Cool down
      } catch (error) {
        return {
          breakingPoint: connections,
          error: error.message,
          status: 'error'
        };
      }
    }

    return {
      breakingPoint: 'none_detected',
      maxTested: Math.max(...stressLevels),
      status: 'passed'
    };
  }

  async runEnduranceTest() {
    console.log('Starting 10-minute endurance test...');

    const startTime = Date.now();
    const result = await this.runLoadTest({
      name: 'endurance',
      connections: 50,
      duration: 600, // 10 minutes
      url: `${this.config.baseUrl}/chat`
    });

    const endTime = Date.now();
    const duration = endTime - startTime;

    return {
      ...result,
      actualDuration: duration,
      stability: this.assessStability(result, duration)
    };
  }

  async runLoadTest(options) {
    const instance = autocannon({
      url: options.url,
      method: options.method || 'GET',
      body: options.body,
      headers: options.headers,
      connections: options.connections,
      duration: options.duration,
      setupClient: this.setupMonitoring.bind(this)
    });

    return new Promise((resolve, reject) => {
      instance.on('done', (results) => {
        resolve({
          timestamp: new Date().toISOString(),
          ...options,
          results: {
            requests: results.requests,
            throughput: results.throughput,
            latency: results.latency,
            errors: results.errors,
            timeouts: results.timeouts
          }
        });
      });

      instance.on('error', reject);
      autocannon.track(instance);
    });
  }

  setupMonitoring(client) {
    // Set up client-side monitoring
    client.on('response', (status, res, context) => {
      this.monitoring.recordResponse(status, res, context);
    });
  }

  summarizeScenarioResults(phaseResults) {
    const totalRequests = phaseResults.reduce((sum, phase) => sum + phase.results.requests.total, 0);
    const totalErrors = phaseResults.reduce((sum, phase) => sum + phase.results.errors, 0);
    const avgLatency = phaseResults.reduce((sum, phase) => sum + phase.results.latency.average, 0) / phaseResults.length;

    return {
      totalRequests,
      totalErrors,
      errorRate: totalErrors / totalRequests,
      avgLatency,
      phasesCount: phaseResults.length
    };
  }

  assessStability(result, duration) {
    const errorRate = result.errors / result.requests.total;
    const latencyVariance = this.calculateLatencyVariance(result);

    let stability = 'excellent';

    if (errorRate > 0.05 || latencyVariance > 1000) {
      stability = 'poor';
    } else if (errorRate > 0.01 || latencyVariance > 500) {
      stability = 'fair';
    } else if (errorRate > 0.001 || latencyVariance > 100) {
      stability = 'good';
    }

    return {
      rating: stability,
      errorRate,
      latencyVariance,
      duration
    };
  }

  calculateLatencyVariance(result) {
    // Simplified variance calculation
    const latencies = [result.latency.p50, result.latency.p95, result.latency.p99];
    const mean = latencies.reduce((a, b) => a + b, 0) / latencies.length;
    const variance = latencies.reduce((sum, latency) => sum + Math.pow(latency - mean, 2), 0) / latencies.length;
    return Math.sqrt(variance);
  }

  generateComprehensiveReport(testResults) {
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        totalTests: testResults.length,
        duration: this.calculateTotalDuration(testResults),
        overallStatus: this.assessOverallStatus(testResults)
      },
      results: testResults,
      recommendations: this.generateRecommendations(testResults),
      performanceMetrics: this.monitoring.getSummary()
    };

    return report;
  }

  calculateTotalDuration(results) {
    return results.reduce((sum, result) => {
      if (result.phases) {
        return sum + result.phases.reduce((phaseSum, phase) => phaseSum + phase.duration, 0);
      }
      return sum + (result.duration || 0);
    }, 0);
  }

  assessOverallStatus(results) {
    const failedTests = results.filter(r => r.status === 'failed' || r.error);
    const errorTests = results.filter(r => r.results?.errors > r.results?.requests?.total * 0.05);

    if (failedTests.length > 0) {
      return 'failed';
    } else if (errorTests.length > results.length * 0.3) {
      return 'degraded';
    } else {
      return 'passed';
    }
  }

  generateRecommendations(results) {
    const recommendations = [];

    const highErrorTests = results.filter(r => r.results?.errors > r.results?.requests?.total * 0.05);
    if (highErrorTests.length > 0) {
      recommendations.push({
        priority: 'high',
        issue: 'High error rates detected',
        tests: highErrorTests.map(t => t.name),
        solution: 'Investigate error handling and resource limits'
      });
    }

    const slowTests = results.filter(r => r.results?.latency?.p95 > 1000);
    if (slowTests.length > 0) {
      recommendations.push({
        priority: 'high',
        issue: 'Slow response times detected',
        tests: slowTests.map(t => t.name),
        solution: 'Optimize caching and database queries'
      });
    }

    return recommendations;
  }

  saveReport(report) {
    const filename = `load-test-report-${new Date().toISOString().slice(0, 10)}.json`;
    require('fs').writeFileSync(filename, JSON.stringify(report, null, 2));
    console.log(`Report saved to ${filename}`);
  }

  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Example usage
const tester = new AdvancedLoadTester({
  baseUrl: 'https://api.yourdomain.com',
  scenarios: [
    {
      name: 'user_journey',
      phases: [
        { name: 'health_check', connections: 10, duration: 30, url: '/health' },
        { name: 'chat_requests', connections: 50, duration: 60, url: '/chat', method: 'POST' },
        { name: 'analytics', connections: 20, duration: 30, url: '/analytics' }
      ]
    }
  ]
});

tester.runComprehensiveTest().then(report => {
  console.log('Load testing completed:', report.summary);
});
```

#### Afternoon: Performance Monitoring and Alerting

```yaml
duration: 3 hours
format: Workshop

objectives:
  - Implement comprehensive performance monitoring
  - Set up performance alerting
  - Create performance dashboards
  - Establish performance baselines

hands-on_exercises:
  1. Performance Monitoring Setup
  2. Alert Configuration
  3. Dashboard Creation
  4. Performance Baseline Establishment
```

#### Exercise 3.2: Performance Monitoring Dashboard

```javascript
// performance-dashboard.js
export class PerformanceDashboard {
  constructor(config = {}) {
    this.config = {
      updateInterval: config.updateInterval || 30000,
      metrics: config.metrics || ['response_time', 'throughput', 'error_rate', 'cache_hit_rate'],
      alerts: config.alerts || {},
      ...config
    };

    this.metrics = new Map();
    this.alerts = new Map();
    this.baselines = new Map();

    this.initializeDashboard();
  }

  initializeDashboard() {
    // Set up real-time updates
    setInterval(() => {
      this.updateMetrics();
      this.checkAlerts();
      this.updateDisplay();
    }, this.config.updateInterval);

    // Initialize baseline values
    this.establishBaselines();

    console.log('Performance dashboard initialized');
  }

  async updateMetrics() {
    try {
      // Fetch current metrics from monitoring system
      const currentMetrics = await this.fetchCurrentMetrics();

      // Store metrics with timestamp
      const timestamp = Date.now();
      for (const [metric, value] of Object.entries(currentMetrics)) {
        if (!this.metrics.has(metric)) {
          this.metrics.set(metric, []);
        }

        const metricData = this.metrics.get(metric);
        metricData.push({ timestamp, value });

        // Keep last 24 hours of data (assuming 30s intervals = 2880 points)
        if (metricData.length > 2880) {
          metricData.shift();
        }
      }
    } catch (error) {
      console.error('Failed to update metrics:', error);
    }
  }

  async fetchCurrentMetrics() {
    // This would integrate with your actual metrics collection system
    const response = await fetch('/metrics');
    const data = await response.json();

    return {
      response_time_p50: data.response_time.p50,
      response_time_p95: data.response_time.p95,
      response_time_p99: data.response_time.p99,
      throughput: data.requests_per_second,
      error_rate: data.error_rate,
      cache_hit_rate: data.cache.hit_rate,
      memory_usage: data.memory.usage_percent,
      cpu_usage: data.cpu.usage_percent
    };
  }

  establishBaselines() {
    // Establish baseline values for comparison
    const baselines = {
      response_time_p95: 500, // 500ms
      throughput: 1000, // 1000 req/sec
      error_rate: 0.01, // 1%
      cache_hit_rate: 0.8, // 80%
      memory_usage: 70, // 70%
      cpu_usage: 60 // 60%
    };

    for (const [metric, value] of Object.entries(baselines)) {
      this.baselines.set(metric, value);
    }
  }

  checkAlerts() {
    const currentMetrics = this.getLatestMetrics();

    for (const [alertName, alertConfig] of Object.entries(this.config.alerts)) {
      const metricValue = currentMetrics[alertConfig.metric];
      const baselineValue = this.baselines.get(alertConfig.metric);

      if (!metricValue || !baselineValue) continue;

      let triggered = false;
      let severity = 'info';

      switch (alertConfig.condition) {
        case 'above':
          triggered = metricValue > alertConfig.threshold;
          severity = metricValue > baselineValue * 1.5 ? 'critical' : 'warning';
          break;
        case 'below':
          triggered = metricValue < alertConfig.threshold;
          severity = metricValue < baselineValue * 0.5 ? 'critical' : 'warning';
          break;
        case 'change':
          const changePercent = Math.abs((metricValue - baselineValue) / baselineValue);
          triggered = changePercent > alertConfig.threshold;
          severity = changePercent > 0.5 ? 'critical' : 'warning';
          break;
      }

      if (triggered) {
        this.triggerAlert({
          name: alertName,
          severity,
          metric: alertConfig.metric,
          value: metricValue,
          threshold: alertConfig.threshold,
          baseline: baselineValue
        });
      }
    }
  }

  triggerAlert(alert) {
    const alertKey = `${alert.name}_${alert.metric}`;

    // Prevent alert spam (don't trigger same alert within 5 minutes)
    const lastTrigger = this.alerts.get(alertKey);
    if (lastTrigger && Date.now() - lastTrigger < 300000) {
      return;
    }

    this.alerts.set(alertKey, Date.now());

    console.log(`🚨 ALERT: ${alert.name} - ${alert.metric}: ${alert.value} (threshold: ${alert.threshold})`);

    // Send alert to configured channels
    this.sendAlert(alert);
  }

  async sendAlert(alert) {
    // Send to Slack
    if (this.config.slackWebhook) {
      await fetch(this.config.slackWebhook, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          text: `🚨 Performance Alert: ${alert.name}`,
          attachments: [{
            color: alert.severity === 'critical' ? 'danger' : 'warning',
            fields: [
              { title: 'Metric', value: alert.metric, short: true },
              { title: 'Current Value', value: alert.value.toString(), short: true },
              { title: 'Threshold', value: alert.threshold.toString(), short: true },
              { title: 'Severity', value: alert.severity, short: true }
            ]
          }]
        })
      });
    }

    // Send to PagerDuty or other alerting systems
    if (this.config.pagerDutyKey && alert.severity === 'critical') {
      await fetch('https://events.pagerduty.com/v2/enqueue', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          routing_key: this.config.pagerDutyKey,
          event_action: 'trigger',
          payload: {
            summary: `Performance Alert: ${alert.name}`,
            severity: alert.severity,
            source: 'performance-dashboard'
          }
        })
      });
    }
  }

  updateDisplay() {
    const metrics = this.getLatestMetrics();
    const status = this.calculateOverallStatus(metrics);

    // Update console display (in production, this would update a web dashboard)
    console.clear();
    console.log('='.repeat(60));
    console.log('🔥 PERFORMANCE DASHBOARD');
    console.log('='.repeat(60));
    console.log(`Status: ${status.overall}`);
    console.log(`Last Update: ${new Date().toLocaleTimeString()}`);
    console.log('');

    console.log('📊 CURRENT METRICS:');
    for (const [metric, value] of Object.entries(metrics)) {
      const baseline = this.baselines.get(metric);
      const status = this.getMetricStatus(metric, value, baseline);
      const displayName = metric.replace(/_/g, ' ').toUpperCase();
      console.log(`${displayName}: ${this.formatValue(metric, value)} ${status.icon} ${status.trend}`);
    }

    console.log('');
    console.log('🚨 ACTIVE ALERTS:');
    for (const [alertKey, timestamp] of this.alerts.entries()) {
      const minutesAgo = Math.floor((Date.now() - timestamp) / 60000);
      console.log(`${alertKey} (${minutesAgo}m ago)`);
    }
  }

  getLatestMetrics() {
    const latest = {};

    for (const [metric, data] of this.metrics.entries()) {
      if (data.length > 0) {
        latest[metric] = data[data.length - 1].value;
      }
    }

    return latest;
  }

  calculateOverallStatus(metrics) {
    let critical = 0;
    let warning = 0;

    for (const [metric, value] of Object.entries(metrics)) {
      const baseline = this.baselines.get(metric);
      const status = this.getMetricStatus(metric, value, baseline);

      if (status.level === 'critical') critical++;
      else if (status.level === 'warning') warning++;
    }

    let overall = 'good';
    if (critical > 0) overall = 'critical';
    else if (warning > 2) overall = 'warning';
    else if (warning > 0) overall = 'fair';

    return { overall, critical, warning };
  }

  getMetricStatus(metric, value, baseline) {
    if (!baseline) return { level: 'unknown', icon: '❓', trend: '' };

    const ratio = value / baseline;
    let level = 'good';
    let icon = '✅';
    let trend = '';

    if (metric.includes('error') || metric.includes('cpu') || metric.includes('memory')) {
      // For these metrics, higher is worse
      if (ratio > 1.5) {
        level = 'critical';
        icon = '🔴';
      } else if (ratio > 1.2) {
        level = 'warning';
        icon = '🟡';
      }
    } else {
      // For performance metrics, lower is worse for response time, higher is better for others
      if (metric.includes('response_time')) {
        if (ratio > 1.5) {
          level = 'critical';
          icon = '🔴';
        } else if (ratio > 1.2) {
          level = 'warning';
          icon = '🟡';
        }
      } else {
        if (ratio < 0.7) {
          level = 'critical';
          icon = '🔴';
        } else if (ratio < 0.85) {
          level = 'warning';
          icon = '🟡';
        }
      }
    }

    // Calculate trend (compare with 5 minutes ago)
    const metricData = this.metrics.get(metric);
    if (metricData && metricData.length > 10) {
      const recent = metricData.slice(-10);
      const older = metricData.slice(-20, -10);
      const recentAvg = recent.reduce((sum, d) => sum + d.value, 0) / recent.length;
      const olderAvg = older.reduce((sum, d) => sum + d.value, 0) / older.length;

      const trendRatio = recentAvg / olderAvg;
      if (trendRatio > 1.05) {
        trend = '↗️';
      } else if (trendRatio < 0.95) {
        trend = '↘️';
      } else {
        trend = '➡️';
      }
    }

    return { level, icon, trend };
  }

  formatValue(metric, value) {
    if (metric.includes('rate') || metric.includes('percent')) {
      return `${(value * 100).toFixed(2)}%`;
    } else if (metric.includes('time')) {
      return `${value.toFixed(0)}ms`;
    } else if (metric.includes('throughput')) {
      return `${value.toFixed(0)}/s`;
    } else {
      return value.toFixed(2);
    }
  }

  getMetricsHistory(metric, hours = 1) {
    const data = this.metrics.get(metric);
    if (!data) return [];

    const cutoff = Date.now() - (hours * 60 * 60 * 1000);
    return data.filter(d => d.timestamp > cutoff);
  }

  exportMetrics(format = 'json') {
    const data = {};

    for (const [metric, values] of this.metrics.entries()) {
      data[metric] = values;
    }

    if (format === 'json') {
      return JSON.stringify(data, null, 2);
    } else if (format === 'csv') {
      // Convert to CSV format
      const timestamps = new Set();
      for (const values of data[Object.keys(data)[0]]) {
        timestamps.add(values.timestamp);
      }

      const sortedTimestamps = Array.from(timestamps).sort();
      let csv = 'timestamp,' + Object.keys(data).join(',') + '\n';

      for (const timestamp of sortedTimestamps) {
        const row = [new Date(timestamp).toISOString()];
        for (const metric of Object.keys(data)) {
          const value = data[metric].find(d => d.timestamp === timestamp);
          row.push(value ? value.value : '');
        }
        csv += row.join(',') + '\n';
      }

      return csv;
    }
  }
}

// Example usage
const dashboard = new PerformanceDashboard({
  alerts: {
    high_error_rate: {
      metric: 'error_rate',
      condition: 'above',
      threshold: 0.05
    },
    slow_response_time: {
      metric: 'response_time_p95',
      condition: 'above',
      threshold: 1000
    },
    low_cache_hit_rate: {
      metric: 'cache_hit_rate',
      condition: 'below',
      threshold: 0.7
    }
  },
  slackWebhook: process.env.SLACK_WEBHOOK,
  pagerDutyKey: process.env.PAGERDUTY_KEY
});

// Export metrics every hour
setInterval(() => {
  const metricsCsv = dashboard.exportMetrics('csv');
  require('fs').writeFileSync(`metrics-${Date.now()}.csv`, metricsCsv);
}, 60 * 60 * 1000);
```

## 📊 Workshop Assessment

### Performance Improvement Metrics

```yaml
assessment_criteria:
  response_time_improvement:
    target: ">30% reduction in P95"
    excellent: ">50% reduction"
    good: "30-50% reduction"
    fair: "10-30% reduction"
    poor: "<10% reduction"

  throughput_improvement:
    target: ">50% increase"
    excellent: ">100% increase"
    good: "50-100% increase"
    fair: "20-50% increase"
    poor: "<20% increase"

  error_rate_reduction:
    target: ">50% reduction"
    excellent: ">80% reduction"
    good: "50-80% reduction"
    fair: "20-50% reduction"
    poor: "<20% reduction"

  cache_hit_rate_improvement:
    target: ">20% increase"
    excellent: ">40% increase"
    good: "20-40% increase"
    fair: "10-20% increase"
    poor: "<10% increase"
```

### Practical Assessment

```yaml
hands_on_assessment:
  duration: 2 hours

  tasks:
    1. Performance Audit:
       - Profile a provided application
       - Identify top 5 performance bottlenecks
       - Document findings with evidence

    2. Optimization Implementation:
       - Implement 3 performance optimizations
       - Measure before/after performance
       - Document optimization rationale

    3. Load Testing:
       - Design and execute a load test
       - Identify breaking points
       - Provide scaling recommendations

    4. Monitoring Setup:
       - Configure performance monitoring
       - Set up alerting for key metrics
       - Create a performance dashboard

  evaluation:
    - Technical accuracy (40%)
    - Optimization effectiveness (30%)
    - Documentation quality (20%)
    - Problem-solving approach (10%)
```

## 📖 Additional Resources

### Tools
- [Autocannon](https://github.com/mcollina/autocannon) - HTTP load testing
- [Artillery](https://artillery.io/) - Advanced load testing
- [Lighthouse](https://developers.google.com/web/tools/lighthouse) - Performance auditing
- [Clinic.js](https://clinicjs.org/) - Performance profiling

### Best Practices
- [Web Performance Optimization](https://web.dev/performance/)
- [Scalability Patterns](https://microservices.io/patterns/)
- [Caching Best Practices](https://redis.io/topics/lru-cache)
- [Load Testing Guide](https://k6.io/docs/)

### Related Documentation

- [Performance Tuning Guide](../operations/performance-tuning.md)
- [Monitoring Setup](../operations/monitoring.md)
- [Architecture Guide](../developer-docs/architecture.md)
- [Extension Guide](../developer-docs/extension-guide.md)
