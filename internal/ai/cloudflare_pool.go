package ai

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CloudflarePool manages multiple Cloudflare worker instances
type CloudflarePool struct {
	providers  []*CloudflareProvider
	current    int
	mu         sync.Mutex
	requests   int64
	lastHealth time.Time
}

// NewCloudflarePool creates a pool of workers
func NewCloudflarePool(workerURLs []string) *CloudflarePool {
	pool := &CloudflarePool{}

	for _, url := range workerURLs {
		provider := NewCloudflareProvider(url)
		pool.providers = append(pool.providers, provider)
	}

	return pool
}

// GetNextProvider returns the next available provider
func (p *CloudflarePool) GetNextProvider() *CloudflareProvider {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Round-robin selection
	provider := p.providers[p.current]
	p.current = (p.current + 1) % len(p.providers)

	return provider
}

// CheckAllHealth checks health of all providers
func (p *CloudflarePool) CheckAllHealth(ctx context.Context) map[string]bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	results := make(map[string]bool)

	for i, provider := range p.providers {
		err := provider.CheckHealth(ctx)
		results[fmt.Sprintf("worker-%d", i)] = err == nil
	}

	p.lastHealth = time.Now()
	return results
}

// GetStats returns pool statistics
func (p *CloudflarePool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_workers":     len(p.providers),
		"current_index":     p.current,
		"total_requests":    p.requests,
		"last_health_check": p.lastHealth,
	}
}
