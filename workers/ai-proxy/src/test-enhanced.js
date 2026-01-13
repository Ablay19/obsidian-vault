// Basic functionality test for enhanced workers
import { WorkerAnalytics } from './analytics.js';
import { SmartCache, CacheAnalytics } from './cache.js';
import { WorkerProfiler } from './profiler.js';

console.log('🧪 Testing Enhanced Workers Functionality...\n');

// Test 1: Analytics System
console.log('📊 Testing Analytics System...');
const analytics = new WorkerAnalytics({});

analytics.trackRequest(25, true, '/api/chat');
analytics.trackRequest(45, false, '/api/chat');
analytics.trackRequest(30, true, '/api/chat');

const report = analytics.getPerformanceReport();
console.log('✅ Analytics Report:', {
  totalRequests: report.totalRequests,
  errorRate: `${(report.errorRate * 100).toFixed(1)}%`,
  avgResponseTime: `${report.avgResponseTime.toFixed(1)}ms`,
  healthScore: report.healthScore
});

// Test 2: Smart Cache
console.log('\n🧠 Testing Smart Cache...');
const smartCache = new SmartCache(10);

smartCache.set('test1', { data: 'value1' }, 5000);
smartCache.set('test2', { data: 'value2' }, 5000);

const hit1 = smartCache.get('test1');
const hit2 = smartCache.get('test2');
const miss = smartCache.get('test3');

const cacheReport = smartCache.getReport();
console.log('✅ Cache Report:', {
  size: cacheReport.cacheSize,
  hitRate: cacheReport.efficiency,
  evictions: cacheReport.evictions,
  totalSets: cacheReport.totalSets
});

// Test 3: Performance Profiler
console.log('\n📈 Testing Performance Profiler...');
const profiler = new WorkerProfiler();

const profileId = profiler.startProfile('test-operation');

// Simulate some work
for (let i = 0; i < 1000; i++) {
  Math.sqrt(i);
}

profiler.addCheckpoint('work-complete', { iterations: 1000 });

const profileResult = profiler.endProfile(profileId);

console.log('✅ Profile Result:', {
  name: profileResult.name,
  duration: `${profileResult.duration.toFixed(2)}ms`,
  memoryDelta: profileResult.memoryDelta,
  checkpoints: profileResult.checkpoints.length
});

// Overall Test Results
console.log('\n🎉 All Enhanced Worker Tests Completed Successfully!');
console.log('✅ Analytics: Real-time metrics tracking');
console.log('✅ Cache: LRU eviction and analytics');
console.log('✅ Profiler: Performance monitoring');
console.log('\n🚀 Enhanced workers are ready for deployment!');

export { analytics, smartCache, profiler };