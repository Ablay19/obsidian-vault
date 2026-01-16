# Tasks: AI Manim Video Generator

**Input**: Design documents from `/specs/006-ai-manim-video/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are NOT included in this task list as they were not explicitly requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Worker project**: `workers/ai-manim-worker/src/`, `workers/ai-manim-worker/tests/`
- **Renderer project**: `workers/manim-renderer/src/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create worker project directory structure at workers/ai-manim-worker/src/{handlers,services,types,utils}
- [ ] T002 Create renderer project directory structure at workers/manim-renderer/src/
- [ ] T003 Initialize TypeScript worker project with package.json including dependencies: @cloudflare/workers-types, hono, wrangler
- [ ] T004 Create tsconfig.json for worker at workers/ai-manim-worker/tsconfig.json
- [ ] T005 Initialize Python renderer project with requirements.txt including: manim, ffmpeg-python, requests
- [ ] T006 [P] Create wrangler.toml configuration at workers/ai-manim-worker/wrangler.toml with KV binding
- [ ] T007 [P] Create Dockerfile for Manim renderer at workers/manim-renderer/Dockerfile with multi-stage build
- [ ] T008 [P] Configure ESLint and Prettier for worker project in .eslintrc.js and .prettierrc

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Type Definitions

- [ ] T009 [P] Define Env interface with environment variables in workers/ai-manim-worker/src/types/index.ts
- [ ] T010 [P] Define ProcessingStatus enum in workers/ai-manim-worker/src/types/index.ts
- [ ] T011 [P] Define UserSession interface in workers/ai-manim-worker/src/types/index.ts
- [ ] T012 [P] Define ProcessingJob interface in workers/ai-manim-worker/src/types/index.ts
- [ ] T013 [P] Define TelegramUpdate interface in workers/ai-manim-worker/src/types/index.ts
- [ ] T014 [P] Define VideoMetadata interface in workers/ai-manim-worker/src/types/index.ts

### Core Infrastructure Services

- [ ] T015 Create structured logger utility in workers/ai-manim-worker/src/utils/logger.ts
- [ ] T016 Implement SessionService for KV operations in workers/ai-manim-worker/src/services/session.ts
- [ ] T017 Implement MockRendererService for development in workers/ai-manim-worker/src/services/mock-renderer.ts
- [ ] T018 Implement AIFallbackService with provider chain in workers/ai-manim-worker/src/services/fallback.ts
- [ ] T019 Create TelegramHandler base structure in workers/ai-manim-worker/src/handlers/telegram.ts
- [ ] T020 Create VideoHandler for video delivery in workers/ai-manim-worker/src/handlers/video.ts
- [ ] T021 Create DebugHandler for health checks in workers/ai-manim-worker/src/handlers/debug.ts
- [ ] T022 Create main Hono app in workers/ai-manim-worker/src/index.ts with route handlers

### Integration Relay Infrastructure (STRONG RELAYS)

- [ ] T023-RELAY Define shared request/response schemas in workers/ai-manim-worker/src/types/api.ts for Worker-Renderer communication
- [ ] T024-RELAY Create Worker-Renderer HTTP client in workers/ai-manim-worker/src/utils/renderer-client.ts with retry logic
- [ ] T025-RELAY Define JobStateSync interface in workers/ai-manim-worker/src/types/sync.ts for KV state synchronization
- [ ] T026-RELAY Implement webhook callback receiver in workers/ai-manim-worker/src/handlers/renderer-callback.ts for renderer completion notifications
- [ ] T027-RELAY Implement error propagation middleware in workers/ai-manim-worker/src/middleware/error-relay.ts for cross-component error handling
- [ ] T028-RELAY Create shared constants in workers/ai-manim-worker/src/utils/constants.ts for timeouts, limits, and status codes

### Renderer Infrastructure

- [ ] T029 Create ManimRendererService interface in workers/ai-manim-worker/src/services/manim.ts
- [ ] T030 Implement Manim rendering script in workers/manim-renderer/src/renderer.py
- [ ] T031 Implement video cleanup script in workers/manim-renderer/src/cleanup.py

### Renderer Integration Points (RELAYS TO WORKER)

- [ ] T032-RELAY Create HTTP server in workers/manim-renderer/src/server.py to accept rendering requests from worker
- [ ] T033-RELAY Implement webhook callback in workers/manim-renderer/src/callback.py to notify worker of render completion
- [ ] T034-RELAY Define worker communication utility in workers/manim-renderer/src/worker-client.py for status updates
- [ ] T035-RELAY Implement shared logging format in workers/manim-renderer/src/logger.py matching worker's structured logger

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Problem Submission (Priority: P1) 🎯 MVP

**Goal**: Enable any Telegram user to submit mathematical problems and receive confirmation with a job ID

**Independent Test**: Any Telegram user can submit problems and receive a job ID confirmation message; invalid requests receive helpful error messages

### Implementation for User Story 1

- [ ] T036 [P] [US1] Implement validateWebhookSecret in workers/ai-manim-worker/src/handlers/telegram.ts to check X-Telegram-Bot-Api-Secret-Token header
- [ ] T037 [P] [US1] Implement parseTelegramUpdate in workers/ai-manim-worker/src/handlers/telegram.ts to extract user ID and message text
- [ ] T038 [P] [US1] Implement createOrGetSession in workers/ai-manim-worker/src/services/session.ts to handle anonymous session creation
- [ ] T039 [P] [US1] Implement createJob in workers/ai-manim-worker/src/services/session.ts to generate ProcessingJob with status=queued
- [ ] T040 [US1] Implement processTelegramMessage in workers/ai-manim-worker/src/handlers/telegram.ts to orchestrate session creation and job creation
- [ ] T041 [US1] Implement validateProblemText in workers/ai-manim-worker/src/handlers/telegram.ts to check 10-5000 character limits and provide helpful error messages
- [ ] T042 [US1] Implement sendTelegramConfirmation in workers/ai-manim-worker/src/handlers/telegram.ts to send job ID confirmation message to user
- [ ] T043 [US1] Implement sendTelegramError in workers/ai-manim-worker/src/handlers/telegram.ts to send helpful error messages for invalid submissions
- [ ] T044 [US1] Add POST /telegram/webhook route in workers/ai-manim-worker/src/index.ts that calls TelegramHandler with secret token validation
- [ ] T045 [US1] Add session auto-extend logic in workers/ai-manim-worker/src/services/session.ts to update last_activity on new submissions
- [ ] T046 [US1] Add logging for problem submission flow in workers/ai-manim-worker/src/handlers/telegram.ts (session creation, job creation, user messaging)

### User Story 1 Integration Relays

- [ ] T047-RELAY [US1] Implement Telegram→Session relay in workers/ai-manim-worker/src/handlers/telegram.ts that passes chat_id to SessionService
- [ ] T048-RELAY [US1] Implement Session→KV relay in workers/ai-manim-worker/src/services/session.ts with automatic TTL refresh on writes
- [ ] T049-RELAY [US1] Implement Session→Telegram confirmation relay in workers/ai-manim-worker/src/handlers/telegram.ts with job_id correlation

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Users can submit problems and receive job IDs.

---

## Phase 4: User Story 2 - AI Problem Solving (Priority: P1)

**Goal**: AI analyzes problems, generates solutions, and produces valid Manim animation code

**Independent Test**: The AI produces Manim code that successfully compiles and generates accurate mathematical visualizations

### Implementation for User Story 2

- [ ] T050 [P] [US2] Implement OpenAIProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [ ] T051 [P] [US2] Implement GeminiProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [ ] T052 [P] [US2] Implement GroqAIProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [ ] T053 [P] [US2] Implement HuggingFaceProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [ ] T054 [P] [US2] Implement DeepSeekProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [ ] T055 [P] [US2] Implement CloudflareAIProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [ ] T056 [US2] Create Manim system prompt in workers/ai-manim-worker/src/services/fallback.ts with Manim v0.18+ syntax requirements
- [ ] T057 [US2] Implement generateManimCode in AIFallbackService to try providers in priority order with error aggregation
- [ ] T058 [US2] Implement code validation in workers/ai-manim-worker/src/services/fallback.ts to check basic Python syntax and required Manim imports
- [ ] T059 [US2] Implement updateJobStatus in workers/ai-manim-worker/src/services/session.ts to update status to ai_generating and save generated code
- [ ] T060 [US2] Add provider health tracking in workers/ai-manim-worker/src/services/fallback.ts to monitor success rates
- [ ] T061 [US2] Integrate AI generation into TelegramHandler in workers/ai-manim-worker/src/handlers/telegram.ts after job creation
- [ ] T062 [US2] Add logging for AI generation flow in workers/ai-manim-worker/src/services/fallback.ts (provider attempts, code length, errors)

### User Story 2 Integration Relays

- [ ] T063-RELAY [US2] Implement Telegram→AI relay in workers/ai-manim-worker/src/handlers/telegram.ts passing problem_text to AIFallbackService
- [ ] T064-RELAY [US2] Implement AI→KV relay in workers/ai-manim-worker/src/services/fallback.ts saving generated_code with job_id correlation
- [ ] T065-RELAY [US2] Implement AI fallback relay in workers/ai-manim-worker/src/services/fallback.ts with provider state tracking across attempts
- [ ] T066-RELAY [US2] Implement AI→Validation relay in workers/ai-manim-worker/src/services/fallback.ts passing code to validator before KV write

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Problems are submitted and AI generates Manim code.

---

## Phase 5: User Story 3 - Video Generation (Priority: P1)

**Goal**: Execute Manim code and produce viewable video files in standard formats

**Independent Test**: Manim code is successfully executed and produces video files within acceptable quality and size parameters

### Implementation for User Story 3

- [ ] T067 [P] [US3] Implement main rendering loop in workers/manim-renderer/src/renderer.py that accepts problem text and code
- [ ] T068 [P] [US3] Implement video output validation in workers/manim-renderer/src/renderer.py to check file size (max 50MB) and format (MP4/WebM)
- [ ] T069 [P] [US3] Implement timeout enforcement in workers/manim-renderer/src/renderer.py to kill renders exceeding 5 minutes
- [ ] T070 [P] [US3] Implement error handling and logging in workers/manim-renderer/src/renderer.py for Manim failures
- [ ] T071 [P] [US3] Implement renderVideo in ManimRendererService in workers/ai-manim-worker/src/services/manim.ts to call renderer container
- [ ] T072 [US3] Implement uploadToR2 in ManimRendererService in workers/ai-manim-worker/src/services/manim.ts to upload generated video to R2 bucket
- [ ] T073 [US3] Implement generatePresignedUrl in ManimRendererService in workers/ai-manim-worker/src/services/manim.ts to create single-access video URL
- [ ] T074 [US3] Implement updateJobWithVideo in workers/ai-manim-worker/src/services/session.ts to update status to ready with video_url and expires_at
- [ ] T075 [US3] Add status transitions to updating in TelegramHandler: code_validating, rendering, uploading
- [ ] T076 [US3] Implement retry logic in TelegramHandler for failed video generation attempts
- [ ] T077 [US3] Add logging for video generation flow in workers/ai-manim-worker/src/handlers/telegram.ts (status updates, render duration, video size)
- [ ] T078 [US3] Configure R2 lifecycle rules via Cloudflare dashboard for 24-hour auto-delete fallback

### User Story 3 Integration Relays (WORKER↔RENDERER)

- [ ] T079-RELAY [US3] Implement Worker→Renderer relay in workers/ai-manim-worker/src/services/manim.ts submitting render requests via HTTP client
- [ ] T080-RELAY [US3] Implement Renderer→Worker relay in workers/manim-renderer/src/callback.py sending completion webhooks with job_id
- [ ] T081-RELAY [US3] Implement Worker→R2 relay in workers/ai-manim-worker/src/services/manim.ts uploading video with metadata tags
- [ ] T082-RELAY [US3] Implement Renderer→R2 direct upload in workers/manim-renderer/src/renderer.py to avoid worker bottleneck
- [ ] T083-RELAY [US3] Implement Worker→KV relay in workers/ai-manim-worker/src/services/session.ts updating job status on renderer callbacks
- [ ] T084-RELAY [US3] Implement status sync relay between Worker and Renderer in workers/ai-manim-worker/src/utils/renderer-client.ts with polling fallback

**Checkpoint**: All three user stories should now be independently functional. Complete flow: problem submission → AI code → video generation.

---

## Phase 6: User Story 4 - Video Delivery (Priority: P2)

**Goal**: Deliver generated videos to users through their interface with immediate deletion after access

**Independent Test**: Users receive videos through their preferred communication channel within acceptable time limits; videos are deleted immediately after access

### Implementation for User Story 4

- [ ] T085 [P] [US4] Implement sendVideoDelivery in workers/ai-manim-worker/src/handlers/telegram.ts to send video URL to user when job status=ready
- [ ] T086 [P] [US4] Implement createVideoLink in workers/ai-manim-worker/src/services/manim.ts to generate accessible web link for video
- [ ] T087 [P] [US4] Implement trackVideoAccess in workers/ai-manim-worker/src/services/session.ts to update job status to delivered when accessed
- [ ] T088 [P] [US4] Implement deleteVideoFromR2 in workers/manim-renderer/src/cleanup.py to immediately delete video after access
- [ ] T089 [P] [US4] Implement updateSessionHistory in workers/ai-manim-worker/src/services/session.ts to add VideoMetadata to video_history array
- [ ] T090 [US4] Implement GET /api/v1/video/{job_id}/access endpoint in workers/ai-manim-worker/src/handlers/video.ts that redirects to presigned URL
- [ ] T091 [US4] Implement immediate video deletion callback after successful delivery in workers/ai-manim-worker/src/handlers/video.ts
- [ ] T092 [US4] Implement GET /api/v1/jobs/{job_id} endpoint in workers/ai-manim-worker/src/handlers/video.ts for job status checking
- [ ] T093 [US4] Implement GET /api/v1/jobs endpoint in workers/ai-manim-worker/src/handlers/video.ts for listing session jobs
- [ ] T094 [US4] Add expiration handling in workers/ai-manim-worker/src/handlers/video.ts to return 410 for already accessed videos
- [ ] T095 [US4] Add mobile-optimized formatting in workers/ai-manim-worker/src/handlers/telegram.ts for video delivery messages
- [ ] T096 [US4] Add logging for video delivery flow in workers/ai-manim-worker/src/handlers/video.ts (delivery confirmation, access tracking, deletion)

### User Story 4 Integration Relays

- [ ] T097-RELAY [US4] Implement KV→Telegram relay in workers/ai-manim-worker/src/handlers/telegram.ts delivering ready jobs to users
- [ ] T098-RELAY [US4] Implement VideoAccess→KV relay in workers/ai-manim-worker/src/handlers/video.ts updating delivered status on access
- [ ] T099-RELAY [US4] Implement KV→R2 deletion relay in workers/ai-manim-worker/src/handlers/video.ts triggering immediate delete on delivery
- [ ] T100-RELAY [US4] Implement R2→SessionHistory relay in workers/ai-manim-worker/src/services/session.ts adding metadata after access
- [ ] T101-RELAY [US4] Implement cross-component access tracking relay in workers/ai-manim-worker/src/services/session.ts correlating job_id across KV, R2, and Telegram

**Checkpoint**: User Story 4 complete. Full end-to-end flow works: problem → AI → video → delivery → deletion.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T102 [P] Update docs/manim-worker.md with deployment instructions and architecture overview
- [ ] T103 [P] Create .env.example file at workers/ai-manim-worker/.env.example with all required environment variables documented
- [ ] T104 [P] Update AGENTS.md with AI provider configuration details from AGENTS.md
- [ ] T105 Implement rate limiting in workers/ai-manim-worker/src/middleware/rate-limit.ts for 10 requests per minute per user
- [ ] T106 Implement request timeout handling in workers/ai-manim-worker/src/utils/timeout.ts for 5-minute max processing
- [ ] T107 Add CORS configuration in workers/ai-manim-worker/src/index.ts for web dashboard access
- [ ] T108 Implement error aggregation in workers/ai-manim-worker/src/utils/errors.ts for better user-facing messages
- [ ] T109 Add structured metrics logging in workers/ai-manim-worker/src/utils/metrics.ts for Cloudflare dashboard
- [ ] T110 Update wrangler.toml with production environment settings
- [ ] T111 Create deployment scripts in workers/ai-manim-worker/scripts/ for worker and renderer deployment
- [ ] T112 Run quickstart.md validation to ensure all setup steps are documented correctly

### Cross-Component Integration Relays

- [ ] T113-RELAY Implement end-to-end data flow verification in workers/ai-manim-worker/tests/integration/e2e-flow.test.ts
- [ ] T114-RELAY Create relay health checks in workers/ai-manim-worker/src/health/relays.ts verifying all component connections
- [ ] T115-RELAY Implement distributed tracing in workers/ai-manim-worker/src/utils/tracing.ts tracking requests across Telegram→Worker→AI→Renderer→R2
- [ ] T116-RELAY Add circuit breakers in workers/ai-manim-worker/src/utils/circuit-breaker.ts for Worker→Renderer communication failures

---

## Integration Relays Architecture

### Strong Component Connections

The project uses explicit relays to ensure strong integration between all components:

```
Telegram Bot ──(Webhook)──> Worker
                              │
                              ├─> KV Store ─┐
                              │               │
                              └─> AI Service ─> Fallback Chain
                                                     │
                              ┌────────────────────┘
                              │
                        Renderer Server
                              │
                              ├─> R2 Storage
                              │
                              └─(Webhook)──> Worker ─> KV Update ─> Telegram (Delivery)
```

### Relay Types

1. **Request Flow Relays**: Telegram → Worker → AI → Renderer → R2
2. **Response Flow Relays**: Renderer → Worker → KV → Telegram
3. **State Synchronization Relays**: Job status updates across KV
4. **Error Propagation Relays**: Cross-component error handling
5. **Data Persistence Relays**: Session tracking across requests

### Key Relay Tasks

| Relay | Task | Components | Purpose |
|-------|------|-----------|---------|
| Telegram→Session | T047-RELAY | TelegramHandler→SessionService | Chat ID to session mapping |
| Session→KV | T048-RELAY | SessionService→KV | Auto-refresh TTL on writes |
| Telegram→AI | T063-RELAY | TelegramHandler→AIFallbackService | Problem text to AI |
| AI→KV | T064-RELAY | AIFallbackService→KV | Save generated code |
| Worker→Renderer | T079-RELAY | ManimRendererService→Renderer Server | Submit render request |
| Renderer→Worker | T080-RELAY | Renderer Callback→Worker | Completion notification |
| Renderer→R2 | T082-RELAY | Renderer→R2 | Direct video upload |
| KV→Telegram | T097-RELAY | SessionService→TelegramHandler | Deliver ready jobs |
| KV→R2 Deletion | T099-RELAY | SessionService→R2 | Immediate delete after access |

### Data Flow Across Relays

```
[User submits problem]
         │
         ▼
Telegram → Worker webhook (T036)
         │
         ▼
Session creation/retrieval (T038) → KV write (T048-RELAY)
         │
         ▼
AI generation (T050-T055) → KV save code (T064-RELAY)
         │
         ▼
Render request (T079-RELAY) → Renderer HTTP
         │
         ▼
Renderer executes (T067-T069) → R2 upload (T082-RELAY)
         │
         ▼
Renderer callback (T080-RELAY) → Worker update (T083-RELAY) → KV
         │
         ▼
Telegram delivery (T085) → Video URL sent
         │
         ▼
[User accesses video] → Access tracking (T087) → KV update
         │
         ▼
Immediate R2 deletion (T099-RELAY)
```

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 (Problem Submission) can start after Foundational - NO dependencies on other stories
  - US2 (AI Problem Solving) can start after Foundational - can work in parallel with US1
  - US3 (Video Generation) can start after US2 (needs AI code)
  - US4 (Video Delivery) can start after US3 (needs generated video)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

```
US1 (Problem Submission) ──────┐
                               ├──> All independent after Foundational
US2 (AI Problem Solving) ──────┘
        │
        └──> US3 (Video Generation) - needs AI code from US2
                │
                └──> US4 (Video Delivery) - needs video from US3
```

### Within Each User Story

- Models before services (if applicable)
- Services before endpoints/handlers
- Core implementation before integration
- Validation and error handling after core logic
- Logging added throughout

### Parallel Opportunities

**Setup Phase (Phase 1)**:
- T006, T007, T008 can run in parallel

**Foundational Phase (Phase 2)**:
- T009-T014 (all type definitions) can run in parallel
- T019-T021 (handler structures) can run in parallel

**User Story 1 (Phase 3)**:
- T036, T037, T038, T039 can run in parallel
- T047-RELAY, T048-RELAY, T049-RELAY can run in parallel after core implementation

**User Story 2 (Phase 4)**:
- T050-T055 (all AI providers) can run in parallel
- T063-RELAY, T064-RELAY, T065-RELAY, T066-RELAY can run in parallel after core implementation

**User Story 3 (Phase 5)**:
- T067, T068, T069, T070 can run in parallel
- T071, T072, T073 can run in parallel
- T079-RELAY, T080-RELAY, T081-RELAY, T082-RELAY, T083-RELAY, T084-RELAY can run in parallel after core implementation

**User Story 4 (Phase 6)**:
- T085, T086, T087, T088, T089 can run in parallel
- T097-RELAY, T098-RELAY, T099-RELAY, T100-RELAY, T101-RELAY can run in parallel after core implementation

**Polish Phase (Phase 7)**:
- T102, T103, T104 can run in parallel
- T113-RELAY, T114-RELAY, T115-RELAY, T116-RELAY can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all core utilities for User Story 1 together:
Task: T036 - Implement validateWebhookSecret in workers/ai-manim-worker/src/handlers/telegram.ts
Task: T037 - Implement parseTelegramUpdate in workers/ai-manim-worker/src/handlers/telegram.ts
Task: T038 - Implement createOrGetSession in workers/ai-manim-worker/src/services/session.ts
Task: T039 - Implement createJob in workers/ai-manim-worker/src/services/session.ts

# After core implementation, launch all integration relays together:
Task: T047-RELAY - Implement Telegram→Session relay in workers/ai-manim-worker/src/handlers/telegram.ts
Task: T048-RELAY - Implement Session→KV relay in workers/ai-manim-worker/src/services/session.ts
Task: T049-RELAY - Implement Session→Telegram confirmation relay in workers/ai-manim-worker/src/handlers/telegram.ts
```

---

## Parallel Example: User Story 2

```bash
# Launch all AI provider implementations for User Story 2 together:
Task: T050 - Implement OpenAIProvider class in workers/ai-manim-worker/src/services/fallback.ts
Task: T051 - Implement GeminiProvider class in workers/ai-manim-worker/src/services/fallback.ts
Task: T052 - Implement GroqAIProvider class in workers/ai-manim-worker/src/services/fallback.ts
Task: T053 - Implement HuggingFaceProvider class in workers/ai-manim-worker/src/services/fallback.ts
Task: T054 - Implement DeepSeekProvider class in workers/ai-manim-worker/src/services/fallback.ts
Task: T055 - Implement CloudflareAIProvider class in workers/ai-manim-worker/src/services/fallback.ts

# After core implementation, launch all integration relays together:
Task: T063-RELAY - Implement Telegram→AI relay in workers/ai-manim-worker/src/handlers/telegram.ts
Task: T064-RELAY - Implement AI→KV relay in workers/ai-manim-worker/src/services/fallback.ts
Task: T065-RELAY - Implement AI fallback relay in workers/ai-manim-worker/src/services/fallback.ts
Task: T066-RELAY - Implement AI→Validation relay in workers/ai-manim-worker/src/services/fallback.ts
```

---

## Parallel Example: User Story 3 (WORKER↔RENDERER RELAYS)

```bash
# Launch all renderer core implementations together:
Task: T067 - Implement main rendering loop in workers/manim-renderer/src/renderer.py
Task: T068 - Implement video output validation in workers/manim-renderer/src/renderer.py
Task: T069 - Implement timeout enforcement in workers/manim-renderer/src/renderer.py
Task: T070 - Implement error handling and logging in workers/manim-renderer/src/renderer.py

# Launch all worker-renderer integration relays together:
Task: T079-RELAY - Implement Worker→Renderer relay in workers/ai-manim-worker/src/services/manim.ts
Task: T080-RELAY - Implement Renderer→Worker relay in workers/manim-renderer/src/callback.py
Task: T081-RELAY - Implement Worker→R2 relay in workers/ai-manim-worker/src/services/manim.ts
Task: T082-RELAY - Implement Renderer→R2 direct upload in workers/manim-renderer/src/renderer.py
Task: T083-RELAY - Implement Worker→KV relay in workers/ai-manim-worker/src/services/session.ts
Task: T084-RELAY - Implement status sync relay between Worker and Renderer in workers/ai-manim-worker/src/utils/renderer-client.ts
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Problem Submission)
4. **STOP and VALIDATE**: Test User Story 1 independently
   - Submit a problem via Telegram
   - Receive job ID confirmation
   - Verify session is created in KV
   - Verify invalid requests get helpful errors
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
   - Submit problem → AI generates Manim code
   - Verify code is saved in job
   - Verify fallback chain works
4. Add User Story 3 → Test independently → Deploy/Demo
   - Full flow: problem → AI → video generation
   - Verify video is uploaded to R2
   - Verify job status updates through states
5. Add User Story 4 → Test independently → Deploy/Demo
   - Complete flow: problem → AI → video → delivery
   - Verify video deletion after access
   - Verify presigned URLs work
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Problem Submission)
   - Developer B: User Story 2 (AI Problem Solving)
   - Developer C: User Story 3 + 4 (Video pipeline)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- **-RELAY** tasks = strong integration points between components (marked with -RELAY suffix)
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **RELAY tasks create strong connections** between: Telegram ↔ Worker ↔ AI ↔ Renderer ↔ R2 ↔ KV
- Worker and renderer are separate projects that must be coordinated via explicit relay points
- Mock renderer should be used for development to avoid Docker dependencies
- R2 lifecycle rules provide fallback for video deletion; immediate deletion via code is primary
- All AI providers have free tiers; configure at least one for functionality
- Session TTL is 7 days with auto-extend; job TTL varies by status
- **Strong relay architecture ensures** data flow visibility, error propagation, and state synchronization across all components
