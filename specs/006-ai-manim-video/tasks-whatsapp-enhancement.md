# Tasks: AI Manim Video Generator - WhatsApp & Direct Code Enhancement

**Input**: Enhancement request for WhatsApp integration and direct Manim code submission
**Prerequisites**: Existing 006-ai-manim-video implementation, plan.md, spec.md, data-model.md, contracts/

**New User Stories Added**:
- User Story 5 - Direct Code Submission (Priority: P1)
- User Story 6 - WhatsApp Integration (Priority: P1)

**Organization**: Tasks are grouped by new user stories with integration points to existing system

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US5, US6)
- Include exact file paths in descriptions

## Path Conventions

- **Worker project**: `workers/ai-manim-worker/src/`, `workers/ai-manim-worker/tests/`
- **Renderer project**: `workers/manim-renderer/src/`

---

## Phase 1: Core Infrastructure for Multi-Platform Support

**Purpose**: Extend existing system to support multiple messaging platforms and direct code submission

### Extend Data Models

- [ ] T001 [P] Add Platform enum to types in workers/ai-manim-worker/src/types/index.ts
- [ ] T002 [P] Update UserSession interface to support multiple platform IDs in workers/ai-manim-worker/src/types/index.ts
- [ ] T003 [P] Extend ProcessingJob interface to include submission_type in workers/ai-manim-worker/src/types/index.ts
- [ ] T004 [P] Add WhatsAppMessage interface to types in workers/ai-manim-worker/src/types/index.ts

### Update Session Service

- [ ] T005 Update SessionService to handle multiple platform IDs in workers/ai-manim-worker/src/services/session.ts
- [ ] T006 Add platform-specific session creation in workers/ai-manim-worker/src/services/session.ts
- [ ] T007 Implement cross-platform session lookup in workers/ai-manim-worker/src/services/session.ts

---

## Phase 2: User Story 5 - Direct Code Submission (Priority: P1) 🎯 MVP Enhancement

**Goal**: Allow users to submit Manim code directly and receive video generation without AI processing

**Independent Test**: Users can submit valid Manim code and receive rendered video; invalid code receives helpful error messages

### Implementation for Direct Code Submission

- [ ] T008 [P] [US5] Add submission_type validation in workers/ai-manim-worker/src/handlers/code.ts
- [ ] T009 [P] [US5] Implement parseCodeSubmission in workers/ai-manim-worker/src/handlers/code.ts to extract Manim code and options
- [ ] T010 [P] [US5] Implement validateManimCode in workers/ai-manim-worker/src/handlers/code.ts with syntax and security checks
- [ ] T011 [P] [US5] Implement createCodeJob in workers/ai-manim-worker/src/services/session.ts for direct code submissions
- [ ] T012 [US5] Implement processCodeSubmission in workers/ai-manim-worker/src/handlers/code.ts to orchestrate direct rendering
- [ ] T013 [US5] Implement sendCodeConfirmation in workers/ai-manim-worker/src/handlers/code.ts for submission acknowledgment
- [ ] T014 [US5] Implement sendCodeError in workers/ai-manim-worker/src/handlers/code.ts for validation failures
- [ ] T015 [US5] Add code submission bypass in TelegramHandler to detect Manim code patterns in workers/ai-manim-worker/src/handlers/telegram.ts
- [ ] T016 [US5] Add logging for code submission flow in workers/ai-manim-worker/src/handlers/code.ts

### Direct Code API Endpoints

- [ ] T017 [P] [US5] Add POST /api/v1/code endpoint in workers/ai-manim-worker/src/index.ts for direct code submission
- [ ] T018 [P] [US5] Add GET /api/v1/code/validate endpoint in workers/ai-manim-worker/src/index.ts for code syntax checking
- [ ] T019 [P] [US5] Update video access endpoint to handle code jobs in workers/ai-manim-worker/src/handlers/video.ts

### Direct Code Integration Relays

- [ ] T020-RELAY [US5] Implement Code→Renderer relay in workers/ai-manim-worker/src/handlers/code.ts bypassing AI generation
- [ ] T021-RELAY [US5] Implement Code→KV relay in workers/ai-manim-worker/src/services/session.ts with submission_type metadata
- [ ] T022-RELAY [US5] Implement Renderer→Status relay for code jobs in workers/ai-manim-worker/src/services/manim.ts

---

## Phase 3: User Story 6 - WhatsApp Integration (Priority: P1) 🎯 Platform Expansion

**Goal**: Enable WhatsApp users to submit problems and code for video generation

**Independent Test**: WhatsApp users can submit problems/code and receive video links through WhatsApp messages

### WhatsApp Webhook Infrastructure

- [ ] T023 [P] [US6] Implement validateWhatsAppWebhook in workers/ai-manim-worker/src/handlers/whatsapp.ts to verify webhook signatures
- [ ] T024 [P] [US6] Implement parseWhatsAppMessage in workers/ai-manim-worker/src/handlers/whatsapp.ts to extract user messages
- [ ] T025 [P] [US6] Implement WhatsAppMessageHandler base structure in workers/ai-manim-worker/src/handlers/whatsapp.ts
- [ ] T026 [US6] Add POST /webhook/whatsapp route in workers/ai-manim-worker/src/index.ts with webhook validation

### WhatsApp Message Processing

- [ ] T027 [P] [US6] Implement detectMessageType in workers/ai-manim-worker/src/handlers/whatsapp.ts for text vs code detection
- [ ] T028 [US6] Implement handleWhatsAppProblem in workers/ai-manim-worker/src/handlers/whatsapp.ts for problem submissions
- [ ] T029 [US6] Implement handleWhatsAppCode in workers/ai-manim-worker/src/handlers/whatsapp.ts for code submissions
- [ ] T030 [US6] Implement sendWhatsAppMessage in workers/ai-manim-worker/src/handlers/whatsapp.ts for user responses
- [ ] T031 [US6] Implement sendWhatsAppVideoLink in workers/ai-manim-worker/src/handlers/whatsapp.ts for video delivery
- [ ] T032 [US6] Implement sendWhatsAppError in workers/ai-manim-worker/src/handlers/whatsapp.ts for error handling

### WhatsApp Session Management

- [ ] T033 [P] [US6] Add WhatsApp session handling in workers/ai-manim-worker/src/services/session.ts with WhatsApp phone numbers
- [ ] T034 [US6] Implement createWhatsAppSession in workers/ai-manim-worker/src/services/session.ts for new users
- [ ] T035 [US6] Add WhatsApp session auto-extend logic in workers/ai-manim-worker/src/services/session.ts

### WhatsApp External Integration

- [ ] T036 [P] [US6] Implement WhatsAppApiClient in workers/ai-manim-worker/src/services/whatsapp.ts for external API calls
- [ ] T037 [P] [US6] Add WhatsApp media upload handling in workers/ai-manim-worker/src/services/whatsapp.ts
- [ ] T038 [P] [US6] Implement WhatsApp webhook verification in workers/ai-manim-worker/src/services/whatsapp.ts
- [ ] T039 [US6] Add WhatsApp rate limiting in workers/ai-manim-worker/src/middleware/rate-limit.ts

### WhatsApp Integration Relays

- [ ] T040-RELAY [US6] Implement WhatsApp→Session relay in workers/ai-manim-worker/src/handlers/whatsapp.ts with phone number mapping
- [ ] T041-RELAY [US6] Implement WhatsApp→AI/Code relay in workers/ai-manim-worker/src/handlers/whatsapp.ts routing to appropriate handler
- [ ] T042-RELAY [US6] Implement Session→WhatsApp relay in workers/ai-manim-worker/src/handlers/whatsapp.ts for message delivery

---

## Phase 4: Platform Abstraction & Cross-Platform Features

**Purpose**: Unify Telegram and WhatsApp handling with shared components

### Abstract Message Handler

- [ ] T043 [P] Create BaseMessageHandler abstract class in workers/ai-manim-worker/src/handlers/base.ts
- [ ] T044 [P] Refactor TelegramHandler to extend BaseMessageHandler in workers/ai-manim-worker/src/handlers/telegram.ts
- [ ] T045 [P] Refactor WhatsAppHandler to extend BaseMessageHandler in workers/ai-manim-worker/src/handlers/whatsapp.ts
- [ ] T046 [P] Create MessageRouter in workers/ai-manim-worker/src/handlers/router.ts to route by platform

### Cross-Platform Features

- [ ] T047 [P] Add platform detection in workers/ai-manim-worker/src/utils/platform-detector.ts
- [ ] T048 [P] Implement unified message formatting in workers/ai-manim-worker/src/utils/message-formatter.ts
- [ ] T049 [P] Add cross-platform job status updates in workers/ai-manim-worker/src/services/session.ts
- [ ] T050 [P] Create unified error handling in workers/ai-manim-worker/src/utils/error-handler.ts

### Enhanced Web Dashboard

- [ ] T051 [P] Add platform selector to web dashboard in workers/ai-manim-worker/public/dashboard.html
- [ ] T052 [P] Implement direct code submission UI in workers/ai-manim-worker/public/dashboard.html with code editor
- [ ] T053 [P] Add WhatsApp connection instructions in workers/ai-manim-worker/public/dashboard.html
- [ ] T054 [P] Update dashboard JavaScript for direct code submission in workers/ai-manim-worker/public/scripts/dashboard.js
- [ ] T055 [P] Add code editor with syntax highlighting in workers/ai-manim-worker/public/scripts/dashboard.js

---

## Phase 5: Testing, Documentation & Polish

**Purpose**: Ensure quality and documentation for new features

### Testing

- [ ] T056 [P] Create unit tests for direct code submission in workers/ai-manim-worker/tests/unit/code.test.ts
- [ ] T057 [P] Create unit tests for WhatsApp handler in workers/ai-manim-worker/tests/unit/whatsapp.test.ts
- [ ] T058 [P] Create integration tests for WhatsApp webhook in workers/ai-manim-worker/tests/integration/whatsapp.test.ts
- [ ] T059 [P] Create integration tests for direct code API in workers/ai-manim-worker/tests/integration/code.test.ts

### Documentation

- [ ] T060 [P] Update API documentation for code endpoints in workers/ai-manim-worker/contracts/openapi.yaml
- [ ] T061 [P] Add WhatsApp integration guide in docs/whatsapp-integration.md
- [ ] T062 [P] Update quickstart guide for multi-platform support in docs/quickstart.md
- [ ] T063 [P] Add direct code submission examples in docs/code-submission.md

### Configuration & Deployment

- [ ] T064 [P] Update environment variables for WhatsApp in workers/ai-manim-worker/.env.example
- [ ] T065 [P] Add WhatsApp configuration to wrangler.toml in workers/ai-manim-worker/wrangler.toml
- [ ] T066 [P] Update deployment scripts for WhatsApp webhook setup in workers/ai-manim-worker/scripts/deploy.sh

---

## Dependencies & Execution Order

### Phase Dependencies

- **Core Infrastructure (Phase 1)**: No dependencies on new features - can start immediately
- **Direct Code Submission (Phase 2)**: Depends on Core Infrastructure completion
- **WhatsApp Integration (Phase 3)**: Depends on Core Infrastructure and Direct Code Submission
- **Platform Abstraction (Phase 4)**: Depends on both Direct Code and WhatsApp implementation
- **Testing & Polish (Phase 5)**: Depends on all feature implementation

### User Story Dependencies

```
US5 (Direct Code) ──────┐
                          ├──> Both independent after Core Infrastructure
US6 (WhatsApp) ────────────┘

US5 (Direct Code) ──────┐
                          └──> US4 (Platform Abstraction) needs both
US6 (WhatsApp) ──────────┘
```

### Parallel Opportunities

**Core Infrastructure (Phase 1)**:
- T001, T002, T003, T004 can run in parallel
- T005, T006, T007 can run in parallel after types are complete

**Direct Code Submission (Phase 2)**:
- T008, T009, T010, T011 can run in parallel
- T017, T018, T019 can run in parallel

**WhatsApp Integration (Phase 3)**:
- T023, T024, T025, T026 can run in parallel
- T027, T028, T029, T030 can run in parallel
- T036, T037, T038, T039 can run in parallel

**Platform Abstraction (Phase 4)**:
- T043, T044, T045 can run in parallel
- T047, T048, T049, T050 can run in parallel

**Testing & Polish (Phase 5)**:
- T056, T057, T058, T059 can run in parallel
- T060, T061, T062, T063 can run in parallel

---

## Implementation Strategy

### MVP Enhancement First (Direct Code Only)

1. Complete Phase 1: Core Infrastructure
2. Complete Phase 2: Direct Code Submission  
3. **STOP and VALIDATE**: Test direct code submission independently
   - Submit Manim code via web dashboard
   - Receive rendered video directly
   - Verify error handling for invalid code
4. Deploy demo if ready

### Incremental Platform Addition

1. Complete Setup + Core Infrastructure → Multi-platform foundation ready
2. Add Direct Code Submission → Test independently → Deploy/Demo
3. Add WhatsApp Integration → Test independently → Deploy/Demo
4. Add Platform Abstraction → Unified system → Deploy/Demo
5. Each addition adds value without breaking existing features

### Parallel Team Strategy

With multiple developers:

1. Team completes Core Infrastructure together
2. Once Core Infrastructure is done:
   - Developer A: Direct Code Submission
   - Developer B: WhatsApp Integration  
   - Developer C: Platform Abstraction (after A&B)
3. Features complete and integrate independently

---

## Key Enhancements Summary

### New Capabilities Added

1. **Direct Manim Code Submission**
   - Users bypass AI generation
   - Direct code validation and rendering
   - Faster turnaround for known code
   - Immediate error feedback

2. **WhatsApp Platform Support**
   - Full webhook integration
   - Message parsing and routing
   - Video delivery via WhatsApp
   - Cross-platform session handling

3. **Unified Multi-Platform System**
   - Abstract message handling
   - Shared session management
   - Consistent error handling
   - Platform-agnostic job processing

### Technical Enhancements

- Extended data models for multi-platform support
- New API endpoints for direct code submission
- Enhanced web dashboard with code editor
- Comprehensive testing coverage
- Updated documentation and deployment guides

---

**Total Task Count: 66**
- Core Infrastructure: 7 tasks
- Direct Code Submission: 15 tasks  
- WhatsApp Integration: 20 tasks
- Platform Abstraction: 8 tasks
- Testing & Polish: 16 tasks

**Independent Test Criteria Per Story**:
- US5: Direct code submission → video rendering or clear error
- US6: WhatsApp users → submit problems/code → receive videos