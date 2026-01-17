# Tasks: Mauritania Network Integration

**Input**: Design documents from `/specs/002-mauritania-net-integration/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included as requested in the specification for quality assurance.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **CLI Package**: `packages/mauritania-net-integration/src/`, `packages/mauritania-net-integration/tests/`
- **Shared Types**: `packages/shared-types/`
- **Contracts**: `specs/002-mauritania-net-integration/contracts/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create Mauritania CLI package structure at cmd/mauritania-cli/ with Go module structure
- [x] T002 Initialize Go project with proper module structure and dependencies (Cobra, Viper, SQLite)
- [x] T003 [P] Create main.go entry point for the CLI application
- [x] T004 [P] Implement root command with Cobra CLI framework
- [x] T005 [P] Create send command for dispatching commands via network transport
- [x] T006 [P] Create status command for checking command execution status
- [x] T007 Create result command for retrieving command execution results

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T008 Implement core type definitions in cmd/mauritania-cli/internal/models/command.go (SocialMediaCommand, NetworkRoute, ShipperSession, CommandResult)
- [x] T009 [P] Create configuration management system in cmd/mauritania-cli/internal/utils/config.go
- [x] T010 [P] Implement SQLite database utilities in cmd/mauritania-cli/internal/database/database.go
- [x] T011 [P] Create encryption utilities for secure credential storage in cmd/mauritania-cli/internal/utils/crypto.go
- [x] T012 [P] Implement authentication validation utilities in cmd/mauritania-cli/internal/utils/auth.go
- [x] T013 [P] Create command parsing and validation utilities in cmd/mauritania-cli/internal/utils/command-parser.go
- [x] T014 [P] Setup internal HTTP server for webhook handling in cmd/mauritania-cli/internal/api/server.go
- [x] T015 Implement offline queue management system in cmd/mauritania-cli/internal/utils/queue.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Social Media Command Interface (Priority: P1) 🎯 MVP

**Goal**: Enable remote command execution through social media APIs for development in low-connectivity regions

**Independent Test**: "User can send commands via social media and receive responses through the same channel, with full shell-like interaction."

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T016 [P] [US1] Contract test for social media API integration in packages/mauritania-net-integration/tests/contract/test-social-media-api.test.ts
- [ ] T017 [P] [US1] Integration test for command send/receive workflow in packages/mauritania-net-integration/tests/integration/test-social-media-commands.test.ts
- [ ] T018 [P] [US1] Integration test for message size limits and pagination in packages/mauritania-net-integration/tests/integration/test-message-limits.test.ts

### Implementation for User Story 1

- [x] T019 [P] [US1] Create WhatsApp transport client in cmd/mauritania-cli/internal/transports/whatsapp/whatsapp.go
- [x] T020 [P] [US1] Create Telegram transport client in cmd/mauritania-cli/internal/transports/telegram/telegram.go
- [x] T021 [P] [US1] Create Facebook transport client in cmd/mauritania-cli/internal/transports/facebook/facebook.go
- [x] T022 [P] [US1] Implement message size handling and pagination in cmd/mauritania-cli/internal/utils/message_handler.go
- [x] T023 [P] [US1] Create rate limiting and throttling utilities in cmd/mauritania-cli/internal/utils/rate_limiter.go
- [x] T024 [US1] Implement social media command receiver in cmd/mauritania-cli/internal/shell/command_receiver.go
- [x] T025 [US1] Create social media response sender in cmd/mauritania-cli/internal/shell/response_sender.go
- [x] T026 [US1] Implement command authentication and validation in cmd/mauritania-cli/internal/services/command_auth.go
- [x] T027 [US1] Add social media transport selection logic in cmd/mauritania-cli/internal/services/transport_selector.go
- [ ] T028 [US1] Implement webhook endpoints for social media callbacks in packages/mauritania-net-integration/src/server/webhooks.ts
- [ ] T029 [US1] Add comprehensive error handling for social media failures in packages/mauritania-net-integration/src/utils/error-handler.ts
- [ ] T030 [US1] Implement command status tracking via social media in packages/mauritania-net-integration/src/services/status-tracker.ts

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - SM APOS Shipper Integration (Priority: P1)

**Goal**: Provide secure command execution through SM APOS Shipper service for Mauritanian developers

**Independent Test**: "Commands are properly routed through SM APOS Shipper service with authentication and secure execution."

### Tests for User Story 2 ⚠️

- [x] T031 [P] [US2] Contract test for SM APOS Shipper API integration in cmd/mauritania-cli/internal/transports/shipper_contract_test.go
- [ ] T032 [P] [US2] Integration test for shipper authentication workflow in cmd/mauritania-cli/internal/integration_test.go
- [ ] T033 [P] [US2] Integration test for secure command execution in cmd/mauritania-cli/internal/integration_test.go
- [x] T034 [P] [US2] Create SM APOS Shipper HTTP client in cmd/mauritania-cli/internal/transports/smapos/smapos.go
- [x] T035 [P] [US2] Implement shipper authentication flow in cmd/mauritania-cli/internal/services/shipper_session_manager.go
- [x] T036 [P] [US2] Create session management for shipper connections in cmd/mauritania-cli/internal/services/shipper_session_manager.go
- [x] T037 [P] [US2] Implement command encryption for shipper transport in cmd/mauritania-cli/internal/utils/command_encryption.go
- [x] T038 [US2] Add shipper command execution engine in cmd/mauritania-cli/internal/services/shipper_command_executor.go
- [x] T039 [US2] Implement shipper result parsing and formatting in cmd/mauritania-cli/internal/utils/shipper_result_parser.go
- [x] T040 [US2] Create shipper rate limiting integration in cmd/mauritania-cli/internal/transports/smapos/smapos.go
- [ ] T041 [US2] Add shipper error handling and retry logic in cmd/mauritania-cli/internal/services/shipper_command_executor.go
- [x] T042 [US2] Implement shipper command validation in cmd/mauritania-cli/internal/services/command_auth.go
- [ ] T043 [US2] Add shipper cost tracking and reporting in cmd/mauritania-cli/internal/services/shipper_command_executor.go
- [ ] T044 [US2] Create shipper session monitoring and health checks in cmd/mauritania-cli/internal/services/shipper_session_manager.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - NRT Network Routing (Priority: P2)

**Goal**: Optimize command execution through intelligent network routing based on cost, reliability, and performance

**Independent Test**: "System automatically selects optimal network routing for command execution based on availability and cost."

### Tests for User Story 3 ⚠️

- [ ] T045 [P] [US3] Contract test for NRT routing API in packages/mauritania-net-integration/tests/contract/test-nrt-api.test.ts
- [ ] T046 [P] [US3] Integration test for route selection algorithm in packages/mauritania-net-integration/tests/integration/test-route-selection.test.ts
- [ ] T047 [P] [US3] Integration test for cost optimization logic in packages/mauritania-net-integration/tests/integration/test-cost-optimization.test.ts

### Implementation for User Story 3

- [ ] T048 [P] [US3] Create NRT routing client in packages/mauritania-net-integration/src/transports/nrt-router.ts
- [ ] T049 [P] [US3] Implement route discovery and enumeration in packages/mauritania-net-integration/src/services/route-discovery.ts
- [ ] T050 [P] [US3] Create route performance monitoring in packages/mauritania-net-integration/src/services/route-monitor.ts
- [ ] T051 [P] [US3] Implement cost calculation algorithms in packages/mauritania-net-integration/src/utils/cost-calculator.ts
- [ ] T052 [US3] Add route selection optimization logic in packages/mauritania-net-integration/src/services/route-optimizer.ts
- [ ] T053 [US3] Implement automatic failover between routes in packages/mauritania-net-integration/src/services/failover-manager.ts
- [ ] T054 [US3] Create route testing and validation utilities in packages/mauritania-net-integration/src/utils/route-tester.ts
- [ ] T055 [US3] Add route preference configuration in packages/mauritania-net-integration/src/services/route-config.ts
- [ ] T056 [US3] Implement route metrics collection and analysis in packages/mauritania-net-integration/src/services/metrics-collector.ts
- [ ] T057 [US3] Create route switching notifications in packages/mauritania-net-integration/src/utils/route-notifier.ts

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T058 [P] Add comprehensive CLI help system and auto-completion
- [ ] T059 [P] Implement mobile-optimized UI and progress indicators
- [ ] T060 [P] Add internationalization support for Arabic/French interfaces
- [ ] T061 [P] Create additional unit tests for edge cases and error conditions
- [ ] T062 [P] Implement comprehensive logging and audit trails
- [ ] T063 [P] Add performance monitoring and optimization features
- [ ] T064 [P] Create backup and restore functionality for configurations
- [ ] T065 [P] Implement plugin system for custom transports
- [ ] T066 [P] Add telemetry and usage analytics (opt-in)
- [ ] T067 [P] Create end-to-end integration test suite
- [ ] T068 [P] Implement CI/CD pipeline integration
- [ ] T069 [P] Add comprehensive documentation and examples
- [ ] T070 [P] Create installation packages for Termux repositories
- [ ] T071 [P] Implement security hardening and penetration testing
- [ ] T072 [P] Add offline mode with local command execution
- [ ] T073 [P] Create user feedback and issue reporting system
- [ ] T074 [P] Implement graceful shutdown and cleanup procedures
- [ ] T075 [P] Add version checking and automatic updates
- [ ] T076 Validate quickstart.md documentation accuracy
- [ ] T077 Run comprehensive cross-platform testing (desktop, Termux, various Android versions)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P1 → P2)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Builds on US1 transport framework but independently testable
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Enhances US1/US2 with routing intelligence

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Transport clients before command execution logic
- Authentication before command processing
- Core functionality before optimization features
- Error handling integrated throughout
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All transport clients within US1 marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members
- All tests for a user story marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Contract test for social media API integration in packages/mauritania-net-integration/tests/contract/test-social-media-api.test.ts"
Task: "Integration test for command send/receive workflow in packages/mauritania-net-integration/tests/integration/test-social-media-commands.test.ts"
Task: "Integration test for message size limits and pagination in packages/mauritania-net-integration/tests/integration/test-message-limits.test.ts"

# Launch all transport clients for User Story 1 together:
Task: "Create WhatsApp transport client in packages/mauritania-net-integration/src/transports/whatsapp.ts"
Task: "Create Telegram transport client in packages/mauritania-net-integration/src/transports/telegram.ts"
Task: "Create Facebook transport client in packages/mauritania-net-integration/src/transports/facebook.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (WhatsApp transport)
4. **STOP and VALIDATE**: Test US1 with real social media commands in Termux
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test with social media → Deploy/Demo (MVP!)
3. Add User Story 2 → Test with SM APOS Shipper → Deploy/Demo
4. Add User Story 3 → Test with NRT routing → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Social Media Interface)
   - Developer B: User Story 2 (SM APOS Shipper)
   - Developer C: User Story 3 (NRT Routing)
3. Stories complete and integrate through shared transport framework

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Focus on Termux compatibility and mobile network conditions
- Social media rate limits and message size constraints are critical
- Authentication and security must be implemented from the start
- Cost optimization is important for Mauritanian mobile users