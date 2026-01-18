# Feature Specification: E2E Testing with Doppler Environment Variables

**Feature Branch**: `002-add-e2e-testing`
**Created**: January 18, 2026
**Status**: Draft
**Input**: end to wnd tests with installing doppler for env vars save in .env or any where

## Clarifications

### Session 2026-01-18

- Q: What is the scope of the E2E testing? → A: Testing for the entire Mauritania CLI project (MCP server, CLI commands, transport integrations, end-to-end user journeys)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Local Testing (Priority: P1)

"As a developer working on the Mauritania CLI, I want end-to-end tests for the entire system (MCP server, CLI commands, transport APIs, user workflows) that use Doppler for environment variables so that I can test with real configurations without exposing secrets in code."

**Why this priority**: Critical for secure development workflow - enables testing with production-like configurations while maintaining security.

**Independent Test**: "Developers can run E2E tests locally with Doppler providing environment variables from secure storage."

**Acceptance Scenarios**:

1. **Given** Doppler is installed and configured, **When** I run E2E tests, **Then** environment variables are automatically injected from Doppler without manual setup
2. **Given** I'm testing transport integrations, **When** I run tests, **Then** API keys and tokens are securely provided by Doppler
3. **Given** I need to test with different environments, **When** I switch Doppler configs, **Then** tests use the appropriate variables

---

### User Story 2 - CI/CD Pipeline Testing (Priority: P1)

"As a DevOps engineer managing the CI/CD pipeline, I want Doppler integration in automated tests so that E2E tests run with proper credentials without storing secrets in repositories."

**Why this priority**: Essential for secure CI/CD - prevents credential leaks while enabling comprehensive testing.

**Independent Test**: "CI/CD pipelines can execute E2E tests with Doppler providing all required environment variables securely."

**Acceptance Scenarios**:

1. **Given** CI/CD pipeline runs E2E tests, **When** Doppler service token is provided, **Then** all environment variables are injected automatically
2. **Given** tests require multiple transport APIs, **When** pipeline executes, **Then** all API credentials are available without being stored in code
3. **Given** a test fails due to missing credentials, **When** Doppler is misconfigured, **Then** clear error messages indicate the issue

---

### User Story 3 - Environment Variable Management (Priority: P2)

"As a tester configuring test environments, I want environment variables saved to .env files or other locations so that I can manage test configurations flexibly and share them with team members."

**Why this priority**: Improves testing flexibility and team collaboration - allows different testing scenarios with appropriate configurations.

**Independent Test**: "Environment variables can be saved to .env files or other configurable locations for different testing scenarios."

**Acceptance Scenarios**:

1. **Given** I need to test with specific configurations, **When** I run Doppler sync, **Then** variables are saved to .env file
2. **Given** team members need the same test setup, **When** they use the .env file, **Then** all tests run with consistent configuration
3. **Given** I need variables in different formats, **When** I configure Doppler, **Then** variables can be exported to .env, JSON, or other formats

---

### Edge Cases

- What happens if Doppler service is unavailable during testing?
- How to handle local development when Doppler is not accessible?
- What if environment variables conflict between Doppler and local .env?
- How to manage different variable sets for different test scenarios?
- What happens if Doppler token expires during long test runs?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST install and configure Doppler CLI for environment variable management
- **FR-002**: System MUST integrate Doppler with E2E test execution to provide environment variables
- **FR-003**: System MUST support saving environment variables to .env files or other configurable locations
- **FR-004**: System MUST handle secure credential injection for transport APIs (WhatsApp, Telegram, etc.)
- **FR-005**: System MUST provide fallback mechanisms when Doppler is unavailable
- **FR-006**: System MUST support multiple Doppler configurations for different testing environments
- **FR-007**: System MUST provide clear error messages for Doppler configuration issues

### Key Entities *(include if feature involves data)*

- **DopplerConfig**: Configuration settings for Doppler CLI and project integration
- **EnvironmentVariable**: Secure key-value pairs managed by Doppler
- **TestEnvironment**: Named configuration sets for different testing scenarios
- **CredentialSet**: Grouped API credentials for transport integrations
- **FallbackConfig**: Alternative variable sources when Doppler is unavailable

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: E2E tests pass in 95% of CI/CD runs with Doppler integration
- **SC-002**: Environment variable setup time reduced from 10 minutes to under 2 minutes
- **SC-003**: Zero credential exposure incidents in test logs or artifacts
- **SC-004**: Tests support 5+ different environment configurations
- **SC-005**: Developer onboarding time for E2E testing reduced by 70%
- **SC-006**: Test environment consistency achieved across all team members
- **SC-007**: Doppler integration adopted in 100% of automated testing workflows

## Assumptions

- Doppler service is available and accessible from development and CI environments
- Team has appropriate Doppler permissions for the project
- Environment variables follow standard naming conventions
- Test frameworks support environment variable injection
- .env file format is acceptable for local development

## Dependencies

- Doppler CLI availability for the target platforms
- Doppler project and config setup
- Test framework compatibility with environment variables
- Network access to Doppler service from test environments

## Out of Scope

- Doppler service setup and account management
- Migration from existing environment variable management
- Integration with specific CI/CD platforms beyond general support
- Advanced Doppler features beyond basic variable injection</content>
<parameter name="filePath">specs/005-architecture-separation/spec.md