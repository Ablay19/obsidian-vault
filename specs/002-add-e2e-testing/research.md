# Research Findings: E2E Testing with Doppler Environment Variables

**Date**: January 18, 2026
**Researcher**: opencode
**Feature**: 002-add-e2e-testing

## Decision: Doppler Integration Approach

**Chosen**: Doppler CLI with service tokens for CI/CD and local development fallback

**Rationale**: Doppler provides secure, centralized environment variable management with CLI tools for seamless integration. Service tokens enable CI/CD automation while local development can use Doppler CLI directly.

**Alternatives Considered**:
- HashiCorp Vault: More complex setup, overkill for testing scenarios
- AWS Secrets Manager: Cloud-specific, additional costs
- Local .env files only: No security for CI/CD

## Decision: E2E Testing Framework

**Chosen**: Custom Go test framework using testify + Doppler integration

**Rationale**: Build on existing testify usage, add Doppler wrapper for environment setup. Allows fine-grained control over test environments and credential injection.

**Alternatives Considered**:
- Cypress: Web-focused, not suitable for CLI testing
- Selenium: Browser automation, not CLI
- Built-in Go testing: Lacks environment management features

## Decision: Fallback Mechanisms

**Chosen**: Multi-tier fallback (Doppler → .env file → default values)

**Rationale**: Ensures tests can run in any environment - production CI uses Doppler, local development uses .env, CI without Doppler uses defaults with warnings.

**Implementation**:
- Primary: Doppler CLI injection
- Secondary: godotenv .env file loading
- Tertiary: Sensible defaults for non-sensitive vars

## Decision: Secure Credential Handling

**Chosen**: Runtime injection with no persistence in test artifacts

**Rationale**: Environment variables are injected at test runtime and never written to logs or files. Doppler service tokens provide temporary access without exposing secrets.

**Security Measures**:
- No credential logging or printing
- Environment cleanup after tests
- Service token rotation in CI/CD
- Audit logging of variable access (without values)

## Decision: Environment Configuration Management

**Chosen**: Doppler configs with environment-specific overrides

**Rationale**: Doppler's config system allows different variable sets per environment (dev, staging, prod) with inheritance and overrides.

**Structure**:
- Base config: Common variables
- Environment configs: Environment-specific overrides
- Test configs: Test-specific variables

## Key Integration Patterns

From research on Doppler CLI usage:

1. **Service Token Authentication**: Use DOPPLER_TOKEN env var for headless operation
2. **Config Selection**: Doppler config names for environment selection
3. **Variable Injection**: `doppler run -- command` or `doppler secrets download`
4. **Local Development**: Doppler CLI login with `doppler login`
5. **CI/CD Integration**: Service tokens in pipeline secrets

## Testing Best Practices

From E2E testing research:

1. **Environment Isolation**: Each test gets clean environment
2. **Parallel Execution**: Tests can run in parallel with isolated env vars
3. **Cleanup**: Environment variables cleaned up after each test
4. **Mock Services**: Use test doubles for external services
5. **Flaky Test Prevention**: Retry logic for network-dependent tests

## Performance Considerations

- Doppler CLI has ~1-2 second startup overhead
- Cache environment variables between test runs
- Parallel test execution with environment isolation
- Lazy loading of variables only when needed

## Error Handling Patterns

- Graceful degradation when Doppler unavailable
- Clear error messages for missing configurations
- Fallback to local .env with warnings
- Timeout handling for Doppler API calls

## Platform Compatibility

- Linux/macOS/Windows support via Doppler CLI
- Go cross-compilation for different architectures
- Docker container support for consistent environments

## Open Questions

- Specific CI/CD platform integration requirements?
- Detailed performance benchmarks for Doppler overhead?
- Integration with existing test reporting tools?