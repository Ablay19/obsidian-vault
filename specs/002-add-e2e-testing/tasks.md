# Tasks: E2E Testing with Doppler Environment Variables

**Input**: Design documents from `/specs/002-add-e2e-testing/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included per TDD requirement (constitution non-negotiable) - tests written first, fail before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- CLI project: `cmd/mauritania-cli/`, `internal/`, `scripts/`
- Tests: `cmd/mauritania-cli/tests/e2e/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Doppler CLI installation and project initialization

- [X] T001 Install Doppler CLI and verify installation
- [X] T002 Create Doppler project structure and environments
- [X] T003 [P] Set up Doppler authentication for development
- [X] T004 Create basic Doppler configuration files

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core Doppler integration framework that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Create Doppler client utilities in internal/doppler/client.go
- [X] T006 Implement environment variable manager in internal/doppler/manager.go
- [X] T007 Add test framework extensions for Doppler in cmd/mauritania-cli/tests/e2e/framework.go
- [X] T008 Create fallback configuration system in internal/doppler/fallback.go
- [X] T009 Implement credential sanitization utilities in internal/doppler/security.go
- [X] T010 Set up Doppler service token management

**Checkpoint**: Doppler integration framework ready - environment variable injection and testing infrastructure can now begin

---

## Phase 3: User Story 1 - Developer Local Testing (Priority: P1) 🎯 MVP

**Goal**: Enable developers to run E2E tests locally with Doppler providing secure environment variables

**Independent Test**: Developers can execute E2E tests locally with Doppler injecting all required credentials without manual setup

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T011 [P] [US1] Unit tests for Doppler client in cmd/mauritania-cli/tests/e2e/doppler_client_test.go
- [ ] T012 [P] [US1] Unit tests for environment manager in cmd/mauritania-cli/tests/e2e/env_manager_test.go
- [ ] T013 [US1] Integration test for local Doppler setup in cmd/mauritania-cli/tests/e2e/local_integration_test.go

### Implementation for User Story 1

- [ ] T014 [P] [US1] Implement local Doppler authentication in internal/doppler/auth.go
- [ ] T015 [P] [US1] Create local test runner with Doppler integration in cmd/mauritania-cli/cmd/test.go
- [ ] T016 [US1] Add Doppler environment injection for test execution
- [ ] T017 [US1] Implement local credential validation and error handling
- [ ] T018 [US1] Create local development documentation

**Checkpoint**: Local Doppler testing functional - developers can run E2E tests with automatic credential injection

---

## Phase 4: User Story 2 - CI/CD Pipeline Testing (Priority: P1)

**Goal**: Enable automated E2E testing in CI/CD pipelines with Doppler service tokens

**Independent Test**: CI/CD pipelines can execute E2E tests with Doppler providing credentials securely without storing secrets

### Tests for User Story 2 ⚠️

- [ ] T019 [P] [US2] Unit tests for service token handling in cmd/mauritania-cli/tests/e2e/service_token_test.go
- [ ] T020 [US2] Integration test for CI/CD Doppler setup in cmd/mauritania-cli/tests/e2e/ci_cd_integration_test.go

### Implementation for User Story 2

- [ ] T021 [P] [US2] Implement service token authentication in internal/doppler/service_tokens.go
- [ ] T022 [US2] Create CI/CD pipeline integration scripts in scripts/ci-doppler-setup.sh
- [ ] T023 [US2] Add automated test execution with Doppler in cmd/mauritania-cli/cmd/ci-test.go
- [ ] T024 [US2] Implement CI/CD environment validation and monitoring
- [ ] T025 [US2] Create CI/CD pipeline documentation and examples

**Checkpoint**: CI/CD Doppler integration complete - automated pipelines can run secure E2E tests

---

## Phase 5: User Story 3 - Environment Variable Management (Priority: P2)

**Goal**: Provide flexible environment variable management with .env file sync and fallback configurations

**Independent Test**: Environment variables can be saved to .env files and managed across different testing scenarios

### Tests for User Story 3 ⚠️

- [ ] T026 [P] [US3] Unit tests for .env file management in cmd/mauritania-cli/tests/e2e/env_file_test.go
- [ ] T027 [US3] Integration test for fallback configurations in cmd/mauritania-cli/tests/e2e/fallback_test.go

### Implementation for User Story 3

- [ ] T028 [P] [US3] Implement .env file sync utilities in internal/doppler/env_sync.go
- [ ] T029 [US3] Create environment configuration management in internal/doppler/config_manager.go
- [ ] T030 [US3] Add multi-environment support (dev/staging/prod)
- [ ] T031 [US3] Implement advanced fallback mechanisms
- [ ] T032 [US3] Create environment management CLI commands

**Checkpoint**: Flexible environment management complete - .env files and fallbacks support all testing scenarios

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T033 [P] Update INSTALLATION.md with Doppler setup instructions
- [ ] T034 Code cleanup and refactoring across Doppler integration
- [ ] T035 Performance optimization for environment loading
- [ ] T036 [P] Additional security tests and validation
- [ ] T037 Create Doppler monitoring and health checks
- [ ] T038 Update AGENTS.md with Doppler technology context
- [ ] T039 Final documentation review and validation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P1 → P2)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Can run parallel to US1
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Benefits from US1/US2 but independently implementable

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Authentication and setup before injection
- Core functionality before advanced features
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit tests for Doppler client in cmd/mauritania-cli/tests/e2e/doppler_client_test.go"
Task: "Unit tests for environment manager in cmd/mauritania-cli/tests/e2e/env_manager_test.go"

# Launch all implementation for User Story 1 together:
Task: "Implement local Doppler authentication in internal/doppler/auth.go"
Task: "Create local test runner with Doppler integration in cmd/mauritania-cli/cmd/test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test local Doppler integration independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Doppler infrastructure ready
2. Add User Story 1 → Local testing with Doppler → Deploy/Demo (MVP!)
3. Add User Story 2 → CI/CD integration → Deploy/Demo
4. Add User Story 3 → Advanced env management → Deploy/Demo
5. Each story adds value without breaking previous functionality

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (local testing)
   - Developer B: User Story 2 (CI/CD integration)
   - Developer C: User Story 3 (env management)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence