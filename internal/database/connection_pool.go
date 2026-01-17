package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// ConnectionPoolConfig holds configuration for database connection pooling
type ConnectionPoolConfig struct {
	DriverName         string        `json:"driver_name" yaml:"driver_name"`
	DSN                string        `json:"dsn" yaml:"dsn"`
	MaxOpenConnections int           `json:"max_open_connections" yaml:"max_open_connections"`
	MaxIdleConnections int           `json:"max_idle_connections" yaml:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime    time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	HealthCheckPeriod  time.Duration `json:"health_check_period" yaml:"health_check_period"`
	RetryAttempts      int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay         time.Duration `json:"retry_delay" yaml:"retry_delay"`
}

// ConnectionPool manages database connections with pooling
type ConnectionPool struct {
	config     ConnectionPoolConfig
	db         *sql.DB
	logger     *zap.Logger
	healthChan chan bool
	stopChan   chan bool
	wg         sync.WaitGroup
	mu         sync.RWMutex
	isHealthy  bool
}

// NewConnectionPool creates a new database connection pool
func NewConnectionPool(config ConnectionPoolConfig, logger *zap.Logger) (*ConnectionPool, error) {
	if config.DriverName == "" {
		return nil, fmt.Errorf("driver name is required")
	}
	if config.DSN == "" {
		return nil, fmt.Errorf("DSN is required")
	}

	// Set default values
	if config.MaxOpenConnections == 0 {
		config.MaxOpenConnections = 25
	}
	if config.MaxIdleConnections == 0 {
		config.MaxIdleConnections = 10
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 5 * time.Minute
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 5 * time.Minute
	}
	if config.HealthCheckPeriod == 0 {
		config.HealthCheckPeriod = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	cp := &ConnectionPool{
		config:     config,
		logger:     logger,
		healthChan: make(chan bool, 1),
		stopChan:   make(chan bool),
		isHealthy:  false,
	}

	// Initialize database connection
	if err := cp.initializeConnection(); err != nil {
		return nil, fmt.Errorf("failed to initialize database connection: %w", err)
	}

	// Start health monitoring
	cp.startHealthMonitoring()

	logger.Info("Database connection pool initialized",
		zap.String("driver", config.DriverName),
		zap.Int("max_open", config.MaxOpenConnections),
		zap.Int("max_idle", config.MaxIdleConnections),
		zap.Duration("max_lifetime", config.ConnMaxLifetime),
	)

	return cp, nil
}

// initializeConnection sets up the database connection with pooling configuration
func (cp *ConnectionPool) initializeConnection() error {
	var err error

	// Retry connection attempts
	for attempt := 1; attempt <= cp.config.RetryAttempts; attempt++ {
		cp.db, err = sql.Open(cp.config.DriverName, cp.config.DSN)
		if err != nil {
			cp.logger.Warn("Failed to open database connection",
				zap.Int("attempt", attempt),
				zap.Error(err))
			if attempt < cp.config.RetryAttempts {
				time.Sleep(cp.config.RetryDelay)
				continue
			}
			return fmt.Errorf("failed to open database after %d attempts: %w", attempt, err)
		}

		// Test the connection
		if err = cp.db.Ping(); err != nil {
			cp.logger.Warn("Failed to ping database",
				zap.Int("attempt", attempt),
				zap.Error(err))
			cp.db.Close()
			if attempt < cp.config.RetryAttempts {
				time.Sleep(cp.config.RetryDelay)
				continue
			}
			return fmt.Errorf("failed to ping database after %d attempts: %w", attempt, err)
		}

		break
	}

	// Configure connection pool
	cp.db.SetMaxOpenConns(cp.config.MaxOpenConnections)
	cp.db.SetMaxIdleConns(cp.config.MaxIdleConnections)
	cp.db.SetConnMaxLifetime(cp.config.ConnMaxLifetime)
	cp.db.SetConnMaxIdleTime(cp.config.ConnMaxIdleTime)

	cp.logger.Info("Database connection pool configured",
		zap.Int("max_open", cp.config.MaxOpenConnections),
		zap.Int("max_idle", cp.config.MaxIdleConnections),
		zap.Duration("max_lifetime", cp.config.ConnMaxLifetime),
		zap.Duration("max_idle_time", cp.config.ConnMaxIdleTime),
	)

	return nil
}

// Get returns the underlying database connection
func (cp *ConnectionPool) Get() *sql.DB {
	return cp.db
}

// Health returns the current health status of the connection pool
func (cp *ConnectionPool) Health() bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.isHealthy
}

// Stats returns connection pool statistics
func (cp *ConnectionPool) Stats() sql.DBStats {
	if cp.db == nil {
		return sql.DBStats{}
	}
	return cp.db.Stats()
}

// Ping tests the database connection
func (cp *ConnectionPool) Ping(ctx context.Context) error {
	if cp.db == nil {
		return fmt.Errorf("database connection not initialized")
	}
	return cp.db.PingContext(ctx)
}

// Close closes the database connection pool
func (cp *ConnectionPool) Close() error {
	cp.logger.Info("Closing database connection pool")

	// Stop health monitoring
	close(cp.stopChan)
	cp.wg.Wait()

	if cp.db != nil {
		return cp.db.Close()
	}
	return nil
}

// startHealthMonitoring starts background health monitoring
func (cp *ConnectionPool) startHealthMonitoring() {
	cp.wg.Add(1)
	go func() {
		defer cp.wg.Done()

		ticker := time.NewTicker(cp.config.HealthCheckPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-cp.stopChan:
				return
			case <-ticker.C:
				cp.performHealthCheck()
			}
		}
	}()
}

// performHealthCheck performs a health check on the database connection
func (cp *ConnectionPool) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cp.Ping(ctx)
	wasHealthy := cp.Health()

	cp.mu.Lock()
	cp.isHealthy = err == nil
	cp.mu.Unlock()

	if err != nil {
		cp.logger.Error("Database health check failed", zap.Error(err))
		if wasHealthy {
			cp.logger.Warn("Database connection became unhealthy")
		}
	} else if !wasHealthy {
		cp.logger.Info("Database connection recovered")
	}

	// Send health status to channel (non-blocking)
	select {
	case cp.healthChan <- cp.isHealthy:
	default:
	}
}

// WaitForHealthy waits for the database to become healthy
func (cp *ConnectionPool) WaitForHealthy(timeout time.Duration) error {
	if cp.Health() {
		return nil
	}

	timeoutChan := time.After(timeout)

	for {
		select {
		case healthy := <-cp.healthChan:
			if healthy {
				return nil
			}
		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for database to become healthy")
		case <-time.After(100 * time.Millisecond):
			if cp.Health() {
				return nil
			}
		}
	}
}

// ExecuteQuery executes a query with proper connection pooling
func (cp *ConnectionPool) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if !cp.Health() {
		return nil, fmt.Errorf("database connection is unhealthy")
	}

	start := time.Now()
	rows, err := cp.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		cp.logger.Error("Database query failed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err))
		return nil, err
	}

	cp.logger.Debug("Database query executed",
		zap.String("query", query),
		zap.Duration("duration", duration))

	return rows, nil
}

// ExecuteStatement executes a statement with proper connection pooling
func (cp *ConnectionPool) ExecuteStatement(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if !cp.Health() {
		return nil, fmt.Errorf("database connection is unhealthy")
	}

	start := time.Now()
	result, err := cp.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		cp.logger.Error("Database statement execution failed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err))
		return nil, err
	}

	cp.logger.Debug("Database statement executed",
		zap.String("query", query),
		zap.Duration("duration", duration))

	return result, nil
}

// BeginTransaction starts a new database transaction
func (cp *ConnectionPool) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	if !cp.Health() {
		return nil, fmt.Errorf("database connection is unhealthy")
	}

	tx, err := cp.db.BeginTx(ctx, nil)
	if err != nil {
		cp.logger.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}

	cp.logger.Debug("Database transaction started")
	return tx, nil
}

// ConnectionPoolManager manages multiple connection pools
type ConnectionPoolManager struct {
	pools  map[string]*ConnectionPool
	logger *zap.Logger
	mu     sync.RWMutex
}

// NewConnectionPoolManager creates a new connection pool manager
func NewConnectionPoolManager(logger *zap.Logger) *ConnectionPoolManager {
	return &ConnectionPoolManager{
		pools:  make(map[string]*ConnectionPool),
		logger: logger,
	}
}

// AddPool adds a new connection pool
func (cpm *ConnectionPoolManager) AddPool(name string, config ConnectionPoolConfig) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if _, exists := cpm.pools[name]; exists {
		return fmt.Errorf("connection pool '%s' already exists", name)
	}

	pool, err := NewConnectionPool(config, cpm.logger.With(zap.String("pool", name)))
	if err != nil {
		return fmt.Errorf("failed to create connection pool '%s': %w", name, err)
	}

	cpm.pools[name] = pool
	cpm.logger.Info("Connection pool added", zap.String("name", name))
	return nil
}

// GetPool returns a connection pool by name
func (cpm *ConnectionPoolManager) GetPool(name string) (*ConnectionPool, error) {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	pool, exists := cpm.pools[name]
	if !exists {
		return nil, fmt.Errorf("connection pool '%s' not found", name)
	}

	return pool, nil
}

// RemovePool removes a connection pool
func (cpm *ConnectionPoolManager) RemovePool(name string) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	pool, exists := cpm.pools[name]
	if !exists {
		return fmt.Errorf("connection pool '%s' not found", name)
	}

	if err := pool.Close(); err != nil {
		cpm.logger.Warn("Error closing connection pool", zap.String("name", name), zap.Error(err))
	}

	delete(cpm.pools, name)
	cpm.logger.Info("Connection pool removed", zap.String("name", name))
	return nil
}

// GetAllPools returns all connection pools
func (cpm *ConnectionPoolManager) GetAllPools() map[string]*ConnectionPool {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	pools := make(map[string]*ConnectionPool)
	for name, pool := range cpm.pools {
		pools[name] = pool
	}
	return pools
}

// HealthCheck performs health checks on all pools
func (cpm *ConnectionPoolManager) HealthCheck() map[string]bool {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	health := make(map[string]bool)
	for name, pool := range cpm.pools {
		health[name] = pool.Health()
	}
	return health
}

// Close closes all connection pools
func (cpm *ConnectionPoolManager) Close() error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	var errors []error
	for name, pool := range cpm.pools {
		if err := pool.Close(); err != nil {
			cpm.logger.Error("Error closing connection pool", zap.String("name", name), zap.Error(err))
			errors = append(errors, fmt.Errorf("pool %s: %w", name, err))
		}
	}

	cpm.pools = make(map[string]*ConnectionPool)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connection pools: %v", errors)
	}
	return nil
}

// DefaultConnectionPoolConfig returns a default configuration for connection pooling
func DefaultConnectionPoolConfig(driver, dsn string) ConnectionPoolConfig {
	return ConnectionPoolConfig{
		DriverName:         driver,
		DSN:                dsn,
		MaxOpenConnections: 25,
		MaxIdleConnections: 10,
		ConnMaxLifetime:    5 * time.Minute,
		ConnMaxIdleTime:    5 * time.Minute,
		HealthCheckPeriod:  30 * time.Second,
		RetryAttempts:      3,
		RetryDelay:         1 * time.Second,
	}
}

// PostgreSQLConnectionPoolConfig returns optimized config for PostgreSQL
func PostgreSQLConnectionPoolConfig(dsn string) ConnectionPoolConfig {
	config := DefaultConnectionPoolConfig("postgres", dsn)
	config.MaxOpenConnections = 20
	config.MaxIdleConnections = 5
	config.ConnMaxLifetime = 10 * time.Minute
	return config
}

// MySQLConnectionPoolConfig returns optimized config for MySQL
func MySQLConnectionPoolConfig(dsn string) ConnectionPoolConfig {
	config := DefaultConnectionPoolConfig("mysql", dsn)
	config.MaxOpenConnections = 25
	config.MaxIdleConnections = 10
	config.ConnMaxLifetime = 5 * time.Minute
	return config
}
