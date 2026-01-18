# Tasks: MCP Integration for Problem Handling

**Input**: Design documents from `/specs/001-add-mcp-integration/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included per TDD requirement (constitution non-negotiable) - tests written first, fail before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Single CLI project: `cmd/mauritania-cli/`, `internal/`
- Tests: `cmd/mauritania-cli/tests/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Add github.com/modelcontextprotocol/go-sdk dependency to go.mod
- [X] T002 Create internal/mcp/ directory structure
- [X] T003 [P] Configure Go linting and formatting tools for MCP package

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core MCP infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Create cmd/mauritania-cli/cmd/mcp-server.go command with transport flags
- [X] T005 Create internal/mcp/server.go with MCP server initialization
- [X] T006 Implement stdio transport support in internal/mcp/transport.go
- [X] T007 Implement HTTP transport support in internal/mcp/transport.go
- [X] T008 Add rate limiting framework in internal/mcp/rate_limiter.go
- [X] T009 Implement data sanitization utilities in internal/mcp/sanitizer.go
- [X] T010 Add session management for MCP connections

**Checkpoint**: Foundation ready - MCP server can start and accept connections, user story tool implementation can now begin in parallel

---

## Phase 3: User Story 1 - AI-Assisted Diagnostics (Priority: P1) 🎯 MVP

**Goal**: Enable AIs to access diagnostic tools for real-time status, logs, and health checks

**Independent Test**: AIs can connect via MCP and successfully query status and logs without manual CLI access

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T011 [P] [US1] Unit tests for status tool in cmd/mauritania-cli/tests/mcp/status_test.go
- [X] T012 [P] [US1] Unit tests for logs tool in cmd/mauritania-cli/tests/mcp/logs_test.go
- [X] T013 [P] [US1] Unit tests for diagnostics tool in cmd/mauritania-cli/tests/mcp/diagnostics_test.go
- [X] T014 [US1] Integration test for MCP server with stdio transport in cmd/mauritania-cli/tests/mcp/server_stdio_test.go

### Implementation for User Story 1

- [X] T015 [P] [US1] Implement status tool in internal/mcp/tools/status.go
- [X] T016 [P] [US1] Implement logs tool in internal/mcp/tools/logs.go
- [X] T017 [P] [US1] Implement diagnostics tool in internal/mcp/tools/diagnostics.go
- [X] T018 [US1] Register tools in server initialization (depends on T015-T017)
- [X] T019 [US1] Add error handling for tool execution failures
- [X] T020 [US1] Add logging for tool usage

**Checkpoint**: At this point, User Story 1 should be fully functional - AIs can diagnose CLI issues via status, logs, and diagnostics tools

---

## Phase 4: User Story 2 - Proactive Problem Resolution (Priority: P1)

**Goal**: Enable AIs to monitor error metrics and test connections proactively

**Independent Test**: AIs can query error metrics, view sanitized configs, and test transport connections independently

### Tests for User Story 2 ⚠️

- [X] T021 [P] [US2] Unit tests for error metrics tool in cmd/mauritania-cli/tests/mcp/metrics_test.go
- [X] T022 [P] [US2] Unit tests for config tool in cmd/mauritania-cli/tests/mcp/config_test.go
- [X] T023 [P] [US2] Unit tests for test connection tool in cmd/mauritania-cli/tests/mcp/test_conn_test.go
- [X] T024 [US2] Integration test for HTTP transport in cmd/mauritania-cli/tests/mcp/server_http_test.go

### Implementation for User Story 2

- [X] T025 [P] [US2] Implement error metrics tool in internal/mcp/tools/metrics.go
- [X] T026 [P] [US2] Implement config tool in internal/mcp/tools/config.go
- [X] T027 [P] [US2] Implement test connection tool in internal/mcp/tools/test_conn.go
- [X] T028 [US2] Register additional tools in server (depends on T025-T027)
- [X] T029 [US2] Add caching for metrics and status data
- [X] T030 [US2] Integrate with existing CLI monitoring infrastructure

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - AIs have full diagnostic and monitoring capabilities

---

## Phase 5: User Story 3 - Multi-AI Collaboration (Priority: P2)

**Goal**: Support multiple AI clients accessing MCP tools simultaneously without conflicts

**Independent Test**: Multiple AIs can concurrently query MCP tools and share diagnostic insights

### Tests for User Story 3 ⚠️

- [X] T031 [P] [US3] Concurrency tests for multiple MCP sessions in cmd/mauritania-cli/tests/mcp/concurrency_test.go
- [X] T032 [US3] Integration test for multiple transport types in cmd/mauritania-cli/tests/mcp/multi_transport_test.go

### Implementation for User Story 3

- [X] T033 [US3] Enhance session management for concurrent access
- [X] T034 [US3] Optimize rate limiting for multiple sessions
- [X] T035 [US3] Add session isolation for data consistency
- [X] T036 [US3] Performance tuning for concurrent tool execution

**Checkpoint**: All user stories should now be independently functional with multi-AI support

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T037 [P] Documentation updates in docs/
- [X] T038 Code cleanup and refactoring across MCP package
- [X] T039 Performance optimization for <2s response times
- [X] T040 [P] Additional unit tests (if requested) in tests/unit/
- [X] T041 Security hardening and vulnerability assessment
- [X] T042 Run quickstart.md validation
- [X] T043 Update AGENTS.md context for MCP technology

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
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - May benefit from US1/US2 testing but independently implementable

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Tools before registration and integration
- Core implementation before optimization
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Tools within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit tests for status tool in cmd/mauritania-cli/tests/mcp/status_test.go"
Task: "Unit tests for logs tool in cmd/mauritania-cli/tests/mcp/logs_test.go"
Task: "Unit tests for diagnostics tool in cmd/mauritania-cli/tests/mcp/diagnostics_test.go"

# Launch all tools for User Story 1 together:
Task: "Implement status tool in internal/mcp/tools/status.go"
Task: "Implement logs tool in internal/mcp/tools/logs.go"
Task: "Implement diagnostics tool in internal/mcp/tools/diagnostics.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (diagnostics tools)
   - Developer B: User Story 2 (monitoring tools)
   - Developer C: User Story 3 (concurrency improvements)
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