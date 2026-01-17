# Feature Flags Documentation

This document describes the feature flag system implemented in the AI Platform for safe, gradual rollout of new features and controlled experimentation.

## 🎯 Overview

Feature flags (also known as feature toggles) allow you to enable or disable features in production without deploying new code. They are essential for:

- **Gradual Rollout**: Release features to subsets of users
- **A/B Testing**: Compare feature performance
- **Canary Deployments**: Test in production safely
- **Emergency Rollback**: Disable problematic features instantly
- **Continuous Deployment**: Deploy code independently of feature release

## 🏗️ Architecture

### Flag Storage

Feature flags are stored in multiple locations depending on use case:

**Static Configuration** (Kubernetes ConfigMaps):
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: feature-flags
  namespace: arch-separation
data:
  USE_NEW_AI_WORKER: "true"
  ENABLE_ADVANCED_LOGGING: "false"
  AI_MODEL_VERSION: "gpt-4"
```

**Dynamic Configuration** (Database/Redis):
```go
type FeatureFlag struct {
    Name        string    `json:"name"`
    Enabled     bool      `json:"enabled"`
    Percentage  int       `json:"percentage"`  // For percentage rollouts
    UserIDs     []string  `json:"user_ids"`    // For user-specific flags
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Flag Evaluation

Flags are evaluated at runtime based on:

- **Global State**: On/off for all users
- **Percentage**: Roll out to X% of users
- **User-specific**: Enable for specific users
- **Environment**: Different behavior per environment

## 🚀 Usage Examples

### Basic Feature Flag

```go
// Check if new AI worker should be used
if featureFlags.UseNewAIWorker {
    return processWithNewAIWorker(request)
} else {
    return processWithLegacyAIWorker(request)
}
```

### Percentage-based Rollout

```go
// Roll out to 10% of users initially
if featureFlags.IsUserInRollout("new-dashboard", userID, 10) {
    return renderNewDashboard(user)
} else {
    return renderOldDashboard(user)
}
```

### Environment-specific Flags

```typescript
// Different behavior per environment
const enableDebugLogging = env === 'development' ||
                          featureFlags.GetBool("ENABLE_DEBUG_LOGGING")

if (enableDebugLogging) {
    logger.debug('Detailed request information', { request, user: userID })
}
```

## 🛠️ Implementation

### Go Implementation

```go
// packages/shared-types/go/features.go
package types

import (
    "crypto/sha256"
    "encoding/hex"
    "strconv"
    "time"
)

type FeatureFlags struct {
    // Static flags from config
    UseNewArchitecture     bool     `json:"use_new_architecture"`
    EnableCircuitBreaker   bool     `json:"enable_circuit_breaker"`
    EnableAdvancedCaching  bool     `json:"enable_advanced_caching"`

    // Dynamic flags
    RolloutPercentage      int      `json:"rollout_percentage"`
    EnabledUserIDs         []string `json:"enabled_user_ids"`
}

type FeatureFlagService interface {
    IsEnabled(flagName string) bool
    IsEnabledForUser(flagName, userID string) bool
    GetPercentageRollout(flagName string, percentage int) bool
    UpdateFlag(flagName string, enabled bool) error
}

// Default implementation
type DefaultFeatureFlagService struct {
    flags map[string]interface{}
}

func (s *DefaultFeatureFlagService) IsEnabled(flagName string) bool {
    if val, exists := s.flags[flagName]; exists {
        if enabled, ok := val.(bool); ok {
            return enabled
        }
    }
    return false
}

func (s *DefaultFeatureFlagService) IsEnabledForUser(flagName, userID string) bool {
    // Check if user is in enabled list
    if users, exists := s.flags[flagName+"_users"]; exists {
        if userList, ok := users.([]string); ok {
            for _, id := range userList {
                if id == userID {
                    return true
                }
            }
        }
    }

    // Check percentage rollout
    if percentage, exists := s.flags[flagName+"_percentage"]; exists {
        if pct, ok := percentage.(int); ok {
            return s.isUserInPercentage(userID, pct)
        }
    }

    return s.IsEnabled(flagName)
}

func (s *DefaultFeatureFlagService) isUserInPercentage(userID string, percentage int) bool {
    // Consistent hashing for user distribution
    hash := sha256.Sum256([]byte(userID))
    hashInt := int(hash[0]) // Use first byte for simplicity
    return hashInt%100 < percentage
}

func (s *DefaultFeatureFlagService) GetPercentageRollout(flagName string, percentage int) bool {
    // Get user ID from context (implementation depends on your auth system)
    userID := getCurrentUserID()
    return s.isUserInPercentage(userID, percentage)
}
```

### TypeScript Implementation

```typescript
// packages/shared-types/typescript/features.ts

export interface FeatureFlags {
    useNewArchitecture: boolean
    enableCircuitBreaker: boolean
    enableAdvancedCaching: boolean
    rolloutPercentage: number
    enabledUserIDs: string[]
}

export class FeatureFlagService {
    private flags: Map<string, any> = new Map()

    constructor(initialFlags: Record<string, any> = {}) {
        Object.entries(initialFlags).forEach(([key, value]) => {
            this.flags.set(key, value)
        })
    }

    isEnabled(flagName: string): boolean {
        return this.flags.get(flagName) === true
    }

    isEnabledForUser(flagName: string, userID: string): boolean {
        // Check user-specific enablement
        const userList = this.flags.get(flagName + '_users')
        if (Array.isArray(userList) && userList.includes(userID)) {
            return true
        }

        // Check percentage rollout
        const percentage = this.flags.get(flagName + '_percentage')
        if (typeof percentage === 'number') {
            return this.isUserInPercentage(userID, percentage)
        }

        return this.isEnabled(flagName)
    }

    private isUserInPercentage(userID: string, percentage: number): boolean {
        // Simple hash-based distribution
        let hash = 0
        for (let i = 0; i < userID.length; i++) {
            hash = ((hash << 5) - hash + userID.charCodeAt(i)) & 0xffffffff
        }
        return Math.abs(hash) % 100 < percentage
    }

    updateFlag(flagName: string, value: any): void {
        this.flags.set(flagName, value)
    }

    getFlag(flagName: string, defaultValue: any = false): any {
        return this.flags.get(flagName) ?? defaultValue
    }
}
```

## 📊 Flag Management

### Flag Definition Schema

```yaml
# config/feature-flags.yaml
flags:
  - name: use_new_ai_worker
    description: "Use the new AI worker implementation"
    type: boolean
    default: false
    environments:
      staging: true
      production: false

  - name: ai_model_rollout
    description: "Roll out new AI model to percentage of users"
    type: percentage
    default: 0
    environments:
      staging: 50
      production: 10

  - name: beta_features
    description: "Enable beta features for specific users"
    type: user_list
    default: []
    environments:
      staging: ["user1", "user2"]
      production: []
```

### Admin Interface

```typescript
// Admin panel for flag management
function FeatureFlagAdmin() {
    const [flags, setFlags] = useState<FeatureFlags>({})

    const updateFlag = async (flagName: string, value: any) => {
        await api.post('/admin/feature-flags', { flagName, value })
        // Refresh flags
        loadFlags()
    }

    return (
        <div>
            <h2>Feature Flag Management</h2>
            {Object.entries(flags).map(([name, value]) => (
                <div key={name}>
                    <label>{name}</label>
                    <input
                        type="checkbox"
                        checked={value as boolean}
                        onChange={(e) => updateFlag(name, e.target.checked)}
                    />
                </div>
            ))}
        </div>
    )
}
```

## 🔄 Rollout Strategies

### 1. Percentage-based Rollout

```go
// Gradual rollout over time
func shouldUseNewFeature(userID string) bool {
    // Day 1-7: 1%
    // Day 8-14: 5%
    // Day 15-21: 25%
    // Day 22+: 100%

    daysSinceRelease := time.Since(featureReleaseDate).Hours() / 24
    percentage := calculateRolloutPercentage(daysSinceRelease)

    return featureFlags.GetPercentageRollout("new_feature", percentage)
}
```

### 2. User-segmented Rollout

```go
// Roll out to beta users first
func shouldUseNewFeature(userID string) bool {
    if isBetaUser(userID) {
        return featureFlags.IsEnabled("new_feature_beta")
    }

    if isPowerUser(userID) {
        return featureFlags.GetPercentageRollout("new_feature_power_users", 50)
    }

    return featureFlags.GetPercentageRollout("new_feature_general", 10)
}
```

### 3. Environment-based Rollout

```go
// Different behavior per environment
func getFeatureConfig() FeatureConfig {
    switch getEnvironment() {
    case "development":
        return FeatureConfig{
            EnableDebugLogging: true,
            UseNewUI: true,
            AITimeout: 30 * time.Second,
        }
    case "staging":
        return FeatureConfig{
            EnableDebugLogging: false,
            UseNewUI: true,
            AITimeout: 20 * time.Second,
        }
    case "production":
        return FeatureConfig{
            EnableDebugLogging: false,
            UseNewUI: featureFlags.IsEnabled("new_ui_rollout"),
            AITimeout: 15 * time.Second,
        }
    }
}
```

## 📈 Monitoring and Analytics

### Flag Usage Tracking

```go
// Track feature flag usage
func trackFeatureUsage(flagName string, userID string, enabled bool) {
    metrics.IncrementCounter("feature_flag_usage", map[string]string{
        "flag_name": flagName,
        "user_id": userID,
        "enabled": strconv.FormatBool(enabled),
    })
}

// Usage analytics
type FeatureUsageStats struct {
    FlagName      string    `json:"flag_name"`
    UsageCount    int       `json:"usage_count"`
    ConversionRate float64  `json:"conversion_rate"`
    LastUsed      time.Time `json:"last_used"`
}
```

### A/B Testing Integration

```go
// A/B test integration
func getVariant(flagName, userID string) string {
    // Use consistent hashing for variant assignment
    hash := hashUserID(userID + flagName)
    variants := []string{"control", "variant_a", "variant_b"}

    return variants[hash % len(variants)]
}

func trackConversion(flagName, userID, variant string, converted bool) {
    metrics.IncrementCounter("feature_conversion", map[string]string{
        "flag_name": flagName,
        "variant": variant,
        "converted": strconv.FormatBool(converted),
    })
}
```

## 🚨 Emergency Controls

### Kill Switches

```go
// Emergency disable all new features
func emergencyDisable() {
    criticalFlags := []string{
        "use_new_architecture",
        "enable_circuit_breaker",
        "new_ai_worker",
    }

    for _, flag := range criticalFlags {
        featureFlags.Disable(flag)
        logger.Warn("Emergency disabled feature flag", "flag", flag)
    }
}
```

### Circuit Breakers

```go
// Integrate with circuit breaker pattern
func processRequest(req *http.Request) error {
    if !featureFlags.IsEnabled("circuit_breaker_protection") {
        return processNormally(req)
    }

    return circuitBreaker.Call(func() error {
        return processWithProtection(req)
    })
}
```

## 🧪 Testing

### Unit Tests

```go
func TestFeatureFlagService(t *testing.T) {
    service := &DefaultFeatureFlagService{
        flags: map[string]interface{}{
            "test_flag": true,
            "percentage_flag_percentage": 50,
        },
    }

    // Test basic enablement
    assert.True(t, service.IsEnabled("test_flag"))
    assert.False(t, service.IsEnabled("disabled_flag"))

    // Test percentage rollout
    enabledCount := 0
    for i := 0; i < 1000; i++ {
        userID := fmt.Sprintf("user%d", i)
        if service.IsEnabledForUser("percentage_flag", userID) {
            enabledCount++
        }
    }

    // Should be approximately 50% ± 5%
    assert.InDelta(t, 500, enabledCount, 50)
}
```

### Integration Tests

```go
func TestFeatureFlagPersistence(t *testing.T) {
    // Test flag changes persist across restarts
    service := NewFeatureFlagService(db)

    // Enable flag
    err := service.UpdateFlag("test_feature", true)
    assert.NoError(t, err)

    // Simulate restart
    newService := NewFeatureFlagService(db)

    // Flag should still be enabled
    assert.True(t, newService.IsEnabled("test_feature"))
}
```

## 📚 Best Practices

### Flag Naming
- Use descriptive, lowercase names: `enable_circuit_breaker`
- Group related flags: `ai_model_gpt4_rollout`
- Include purpose: `new_ui_beta_users`

### Flag Lifecycle
1. **Create**: Add flag with default false
2. **Develop**: Implement feature with flag control
3. **Test**: Enable in staging, test thoroughly
4. **Rollout**: Gradual percentage increase
5. **Monitor**: Watch metrics and error rates
6. **Stabilize**: Enable for all users
7. **Cleanup**: Remove flag and dead code

### Documentation
- Document each flag's purpose and impact
- Update rollout plans with flag changes
- Maintain flag inventory for compliance

## 🔧 Configuration

### Environment Variables

```bash
# Feature flags via environment
export FEATURE_USE_NEW_ARCHITECTURE=true
export FEATURE_AI_MODEL_ROLLOUT_PERCENTAGE=25
export FEATURE_BETA_USERS="user1,user2,user3"
```

### Configuration Files

```yaml
# config/feature-flags.yaml
feature_flags:
  use_new_architecture: true
  ai_model_rollout_percentage: 25
  beta_users:
    - user1
    - user2
    - user3
```

### Database Storage

```sql
CREATE TABLE feature_flags (
    name VARCHAR(255) PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT false,
    percentage INTEGER DEFAULT 0,
    user_ids JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## 📊 Metrics and Monitoring

### Flag Usage Metrics

```prometheus
# Feature flag usage
feature_flag_usage_total{flag_name, enabled} counter

# Feature flag evaluation time
feature_flag_evaluation_duration_seconds{flag_name} histogram

# A/B test conversion rates
feature_conversion_rate{flag_name, variant} gauge
```

### Alerting Rules

```yaml
# Alert on flag evaluation errors
- alert: FeatureFlagEvaluationError
  expr: rate(feature_flag_evaluation_errors_total[5m]) > 0
  labels:
    severity: warning

# Alert on high flag usage (potential performance impact)
- alert: HighFeatureFlagUsage
  expr: rate(feature_flag_usage_total[5m]) > 1000
  labels:
    severity: info
```

## 🔄 Integration with CI/CD

### Automated Testing

```yaml
# .github/workflows/feature-flag-test.yml
name: Feature Flag Tests
on:
  push:
    paths:
      - 'packages/shared-types/**'
      - '**/feature-flags/**'

jobs:
  test-flags:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Test feature flags
        run: go test ./packages/shared-types/go/... -run TestFeatureFlag
```

### Deployment Integration

```bash
# Deploy with feature flags
kubectl apply -f deploy/k8s/feature-flags.yaml
kubectl rollout restart deployment/api-gateway
```

## 📋 Checklist

### Before Adding a Flag
- [ ] Purpose clearly defined
- [ ] Rollout strategy planned
- [ ] Cleanup plan documented
- [ ] Tests written
- [ ] Monitoring configured

### During Rollout
- [ ] Flag starts disabled
- [ ] Gradual percentage increase
- [ ] Metrics monitored
- [ ] Rollback plan ready
- [ ] User feedback collected

### After Stabilization
- [ ] Flag removed from code
- [ ] Dead code cleaned up
- [ ] Documentation updated
- [ ] Tests updated

Feature flags are powerful tools for safe deployments and experimentation. Use them wisely to reduce risk and enable faster iteration.