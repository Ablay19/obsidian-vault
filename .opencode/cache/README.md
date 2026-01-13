# Opencode Cache Management

This directory manages cached data, optimization, and performance improvements.

## Cache Structure

### Cache Types
- **File Cache** - Cached file contents and metadata
- **Search Cache** - Results from grep and glob operations
- **API Cache** - External API responses and data
- **AI Cache** - AI model responses and completions
- **Compilation Cache** - Build and compilation results
- **Metadata Cache** - File system metadata and permissions

### Cache Organization
```
cache/
├── files/              # File content cache
├── searches/           # Search results cache
├── api/               # External API cache
├── ai/                # AI model cache
├── compiled/          # Compilation/build cache
├── metadata/          # File metadata cache
└── temp/              # Temporary cache files
```

## Cache Configuration

### Settings (`config/cache.yaml`)
```yaml
enabled: true
directory: "./.opencode/cache"
max_size: "1GB"
cleanup_interval: 3600  # seconds

cache_types:
  files:
    enabled: true
    max_size: "500MB"
    ttl: 3600
    compression: true
    
  searches:
    enabled: true
    max_size: "100MB"
    ttl: 1800
    
  api:
    enabled: true
    max_size: "200MB"
    ttl: 300
    
  ai:
    enabled: true
    max_size: "300MB"
    ttl: 86400
    compression: true
    
  compiled:
    enabled: true
    max_size: "100MB"
    ttl: 7200
    
  metadata:
    enabled: true
    max_size: "50MB"
    ttl: 300

policies:
  eviction: "lru"         # lru, lfu, fifo
  compression: "gzip"      # gzip, brotli, none
  encryption: false        # Encrypt sensitive cache
  
performance:
  indexing: true           # Maintain cache index
  prefetching: true        # Predictive cache loading
  background_cleanup: true # Cleanup in background
  parallel_access: true    # Concurrent cache operations
```

## Cache Operations

### Reading from Cache
```javascript
// Get cached file content
const cached = await cache.get('files', {
  path: 'src/main.js',
  checksum: 'abc123',
  modified: '2026-01-13T10:30:00Z'
});

if (cached && !cached.expired) {
  return cached.content;
}

// Cache miss - read from file system
const content = await fs.readFile('src/main.js');
await cache.set('files', {
  path: 'src/main.js',
  checksum: 'abc123',
  content: content,
  timestamp: Date.now()
});

return content;
```

### Writing to Cache
```javascript
// Cache search results
await cache.set('searches', {
  query: 'function test',
  pattern: '**/*.js',
  results: searchResults,
  timestamp: Date.now(),
  ttl: 1800
});

// Cache API response
await cache.set('api', {
  endpoint: 'https://api.example.com/data',
  method: 'GET',
  response: apiResponse,
  timestamp: Date.now(),
  ttl: 300
});
```

### Cache Invalidation
```javascript
// Invalidate specific cache entry
await cache.invalidate('files', 'src/main.js');

// Invalidate by pattern
await cache.invalidate('searches', 'function test*');

// Invalidate entire cache type
await cache.clear('ai');

// Invalidate expired entries
await cache.cleanup();
```

## Cache Keys and Indexing

### Key Generation
```javascript
// File cache key
const fileKey = `file:${path}:${checksum}`;

// Search cache key
const searchKey = `search:${pattern}:${query_hash}`;

// API cache key
const apiKey = `api:${endpoint}:${method}:${params_hash}`;

// AI cache key
const aiKey = `ai:${model}:${prompt_hash}`;
```

### Cache Index
Maintain efficient lookup index:
```json
{
  "files": {
    "src/main.js": {
      "key": "file:src/main.js:abc123",
      "size": 1024,
      "last_access": "2026-01-13T10:30:00Z",
      "access_count": 5
    }
  },
  "searches": {
    "function test": {
      "key": "search:function_test:def456",
      "results_count": 15,
      "last_access": "2026-01-13T10:25:00Z"
    }
  }
}
```

## Cache Performance

### Metrics Tracked
- **Hit Rate** - Percentage of cache hits vs misses
- **Eviction Rate** - How often items are evicted
- **Memory Usage** - Current cache memory consumption
- **Disk Usage** - Cache size on disk
- **Access Patterns** - Most/least accessed items

### Performance Optimization
```javascript
// Predictive caching based on usage patterns
cache.predictiveLoading({
  enabled: true,
  threshold: 0.8,  // Load when 80% likely to be used
  max_prefetch: 10,
  background: true
});

// Adaptive TTL based on usage
cache.adaptiveTTL({
  base_ttl: 3600,
  multiplier: 2.0,  // Double TTL for frequently accessed
  min_ttl: 300,
  max_ttl: 86400
});
```

## Cache Management Commands

### CLI Operations
```bash
# View cache statistics
opencode cache stats

# Clear specific cache type
opencode cache clear ai

# Clean up expired entries
opencode cache cleanup

# Warm up cache with common files
opencode cache warm --project ./src

# Analyze cache performance
opencode cache analyze

# Export cache data
opencode cache export --format json --output cache.json

# Import cache data
opencode cache import --file cache.json
```

### Cache Monitoring
```bash
# Real-time cache monitoring
opencode cache monitor

# Memory usage by cache type
opencode cache memory --type files

# Access patterns analysis
opencode cache patterns --last 24h
```

## Cache Security

### Data Protection
- **Encryption** - Optional encryption for sensitive data
- **Access Control** - Permissions-based cache access
- **Sanitization** - Remove sensitive data before caching
- **Audit Logging** - Track cache access and modifications

### Secure Cache Keys
```javascript
// Generate secure cache keys
const secureKey = crypto.createHash('sha256')
  .update(`${path}:${salt}`)
  .digest('hex');
```

### Cache Isolation
- User-specific cache directories
- Permission-based cache segregation
- Sandboxed cache environments
- Multi-tenant cache support

## Cache Distribution

### Multi-Instance Coordination
```yaml
distributed_cache:
  enabled: true
  backend: "redis"     # redis, file, memory
  endpoint: "redis://localhost:6379"
  
  sync_strategy: "eventual"  # immediate, eventual, manual
  conflict_resolution: "last_writer_wins"
  
  replication:
    enabled: true
    factor: 3
    consistency: "eventual"
```

### Cache Synchronization
```javascript
// Sync cache across instances
await cache.sync({
  remote: 'redis://cache-server:6379',
  strategy: 'bidirectional',
  conflict: 'merge'
});

// Subscribe to cache updates
cache.subscribe('files', (key, value) => {
  console.log(`Cache updated: ${key} = ${value}`);
});
```

## Cache Troubleshooting

### Common Issues

#### **Cache Corruption**
```bash
# Detect corrupted cache entries
opencode cache verify --type files

# Repair corrupted cache
opencode cache repair --type files

# Rebuild cache from scratch
opencode cache rebuild --type files
```

#### **Memory Leaks**
```bash
# Monitor memory usage
opencode cache monitor memory

# Force cleanup
opencode cache cleanup --force

# Reduce cache size
opencode cache resize --max-size 500MB
```

#### **Performance Degradation**
```bash
# Analyze access patterns
opencode cache analyze --patterns

# Optimize cache configuration
opencode cache optimize --aggressive

# Reset cache statistics
opencode cache reset-stats
```

## Cache Analytics

### Performance Reports
```bash
# Generate performance report
opencode cache report --type performance --last 7d

# Hit rate analysis
opencode cache report --type hitrate --last 30d

# Usage patterns
opencode cache report --type patterns --last 24h
```

### Optimization Recommendations
- Identify frequently accessed items for preloading
- Suggest TTL adjustments based on usage patterns
- Recommend cache size adjustments
- Propose configuration changes for better performance