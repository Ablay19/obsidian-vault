# Tasks: AI Manim Video Generator with WhatsApp & Direct Code Enhancement

**Input**: Enhanced feature specification including WhatsApp integration and direct Manim code submission
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**New User Stories Added**:
- User Story 5 - Direct Code Submission (Priority: P1)
- User Story 6 - WhatsApp Integration (Priority: P1)

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

- [x] T001 Create worker project directory structure at workers/ai-manim-worker/src/{handlers,services,types,utils}
- [x] T002 Create renderer project directory structure at workers/manim-renderer/src/
- [x] T003 Initialize TypeScript worker project with package.json including dependencies: @cloudflare/workers-types, hono, wrangler
- [x] T004 Create tsconfig.json for worker at workers/ai-manim-worker/tsconfig.json
- [x] T005 Initialize Python renderer project with requirements.txt including: manim, ffmpeg-python, requests
- [x] T006 [P] Create wrangler.toml configuration at workers/ai-manim-worker/wrangler.toml with KV binding
- [x] T007 [P] Create Dockerfile for Manim renderer at workers/manim-renderer/Dockerfile with multi-stage build
- [x] T008 [P] Configure ESLint and Prettier for worker project in .eslintrc.js and .prettierrc

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Type Definitions

- [x] T009 [P] Define Env interface with environment variables in workers/ai-manim-worker/src/types/index.ts
- [x] T010 [P] Define ProcessingStatus enum in workers/ai-manim-worker/src/types/index.ts
- [x] T011 [P] Define UserSession interface in workers/ai-manim-worker/src/types/index.ts
- [x] T012 [P] Define ProcessingJob interface in workers/ai-manim-worker/src/types/index.ts
- [x] T013 [P] Define TelegramUpdate interface in workers/ai-manim-worker/src/types/index.ts
- [x] T014 [P] Define VideoMetadata interface in workers/ai-manim-worker/src/types/index.ts

### Core Infrastructure Services

- [x] T015 Create structured logger utility in workers/ai-manim-worker/src/utils/logger.ts
- [x] T016 Implement SessionService for KV operations in workers/ai-manim-worker/src/services/session.ts
- [x] T017 Implement MockRendererService for development in workers/ai-manim-worker/src/services/mock-renderer.ts
- [x] T018 Implement AIFallbackService with provider chain in workers/ai-manim-worker/src/services/fallback.ts
- [x] T019 Create TelegramHandler base structure in workers/ai-manim-worker/src/handlers/telegram.ts
- [x] T020 Create VideoHandler for video delivery in workers/ai-manim-worker/src/handlers/video.ts
- [x] T021 Create DebugHandler for health checks in workers/ai-manim-worker/src/handlers/debug.ts
- [x] T022 Create main Hono app in workers/ai-manim-worker/src/index.ts with route handlers

### Integration Relay Infrastructure (STRONG RELAYS)

- [x] T023-RELAY Define shared request/response schemas in workers/ai-manim-worker/src/types/api.ts for Worker-Renderer communication
- [x] T025-RELAY Define JobStateSync interface in workers/ai-manim-worker/src/types/sync.ts for KV state synchronization
- [x] T030 Implement Manim rendering script in workers/manim-renderer/src/renderer.py
- [x] T031 Implement video cleanup script in workers/manim-renderer/src/cleanup.py

### Renderer Integration Points (RELAYS TO WORKER)

- [x] T032-RELAY Create HTTP server in workers/manim-renderer/src/server.py to accept rendering requests from worker
- [x] T033-RELAY Implement webhook callback in workers/manim-renderer/src/callback.py to notify worker of render completion
- [x] T034-RELAY Define worker communication utility in workers/manim-renderer/src/worker-client.py for status updates
- [x] T035-RELAY Implement shared logging format in workers/manim-renderer/src/logger.py matching worker's structured logger

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Problem Submission (Priority: P1) 🎯 MVP

**Goal**: Enable any Telegram user to submit mathematical problems and receive confirmation with a job ID

**Independent Test**: Any Telegram user can submit problems and receive a job ID confirmation message; invalid requests receive helpful error messages

### Implementation for User Story 1

- [x] T036 [P] [US1] Implement validateWebhookSecret in workers/ai-manim-worker/src/handlers/telegram.ts to check X-Telegram-Bot-Api-Secret-Token header
- [x] T037 [P] [US1] Implement parseTelegramUpdate in workers/ai-manim-worker/src/handlers/telegram.ts to extract user ID and message text
- [x] T038 [P] [US1] Implement createOrGetSession in workers/ai-manim-worker/src/services/session.ts to handle anonymous session creation
- [x] T039 [P] [US1] Implement createJob in workers/ai-manim-worker/src/services/session.ts to generate ProcessingJob with status=queued
- [x] T040 [US1] Implement processTelegramMessage in workers/ai-manim-worker/src/handlers/telegram.ts to orchestrate session creation and job creation
- [x] T041 [US1] Implement validateProblemText in workers/ai-manim-worker/src/handlers/telegram.ts to check 10-5000 character limits and provide helpful error messages
- [x] T042 [US1] Implement sendTelegramConfirmation in workers/ai-manim-worker/src/handlers/telegram.ts to send job ID confirmation message to user
- [x] T043 [US1] Implement sendTelegramError in workers/ai-manim-worker/src/handlers/telegram.ts to send helpful error messages for invalid submissions
- [x] T044 [US1] Add POST /telegram/webhook route in workers/ai-manim-worker/src/index.ts that calls TelegramHandler with secret token validation
- [x] T045 [US1] Add session auto-extend logic in workers/ai-manim-worker/src/services/session.ts to update last_activity on new submissions
- [x] T046 [US1] Add logging for problem submission flow in workers/ai-manim-worker/src/handlers/telegram.ts (session creation, job creation, user messaging)

### User Story 1 Integration Relays

- [x] T047-RELAY [US1] Implement Telegram→Session relay in workers/ai-manim-worker/src/handlers/telegram.ts that passes chat_id to SessionService
- [x] T048-RELAY [US1] Implement Session→KV relay in workers/ai-manim-worker/src/services/session.ts with automatic TTL refresh on writes
- [x] T049-RELAY [US1] Implement Session→Telegram confirmation relay in workers/ai-manim-worker/src/handlers/telegram.ts with job_id correlation

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Users can submit problems and receive job IDs.

---

## Phase 4: User Story 2 - AI Problem Solving (Priority: P1)

**Goal**: AI analyzes problems, generates solutions, and produces valid Manim animation code

**Independent Test**: The AI produces Manim code that successfully compiles and generates accurate mathematical visualizations

### Implementation for User Story 2

- [x] T050 [P] [US2] Implement OpenAIProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [x] T051 [P] [US2] Implement GeminiProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [x] T052 [P] [US2] Implement GroqAIProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [x] T053 [P] [US2] Implement HuggingFaceProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [x] T054 [P] [US2] Implement DeepSeekProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [x] T055 [P] [US2] Implement CloudflareAIProvider class in workers/ai-manim-worker/src/services/fallback.ts with generateManimCode method
- [x] T056 [US2] Create Manim system prompt in workers/ai-manim-worker/src/services/fallback.ts with Manim v0.18+ syntax requirements
- [x] T057 [US2] Implement generateManimCode in AIFallbackService to try providers in priority order with error aggregation
- [x] T058 [US2] Implement code validation in workers/ai-manim-worker/src/services/fallback.ts to check basic Python syntax and required Manim imports
- [x] T059 [US2] Implement updateJobStatus in workers/ai-manim-worker/src/services/session.ts to update status to ai_generating and save generated code
- [x] T060 [US2] Add provider health tracking in workers/ai-manim-worker/src/services/fallback.ts to monitor success rates
- [x] T061 [US2] Integrate AI generation into TelegramHandler in workers/ai-manim-worker/src/handlers/telegram.ts after job creation
- [x] T062 [US2] Add logging for AI generation flow in workers/ai-manim-worker/src/services/fallback.ts (provider attempts, code length, errors)

### User Story 2 Integration Relays

- [x] T063-RELAY [US2] Implement Telegram→AI relay in workers/ai-manim-worker/src/handlers/telegram.ts passing problem_text to AIFallbackService
- [x] T064-RELAY [US2] Implement AI→KV relay in workers/ai-manim-worker/src/services/fallback.ts saving generated_code with job_id correlation
- [x] T065-RELAY [US2] Implement AI fallback relay in workers/ai-manim-worker/src/services/fallback.ts with provider state tracking across attempts
- [x] T066-RELAY [US2] Implement AI→Validation relay in workers/ai-manim-worker/src/services/fallback.ts passing code to validator before KV write

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Problems are submitted and AI generates Manim code.

---

## Phase 5: User Story 3 - Video Generation (Priority: P1)

**Goal**: Execute Manim code and produce viewable video files in standard formats

**Independent Test**: Manim code is successfully executed and produces video files within acceptable quality and size parameters

### Implementation for User Story 3

- [x] T067 [P] [US3] Implement main rendering loop in workers/manim-renderer/src/renderer.py that accepts problem text and code
- [x] T068 [P] [US3] Implement video output validation in workers/manim-renderer/src/renderer.py to check file size (max 50MB) and format (MP4/WebM)
- [x] T069 [P] [US3] Implement timeout enforcement in workers/manim-renderer/src/renderer.py to kill renders exceeding 5 minutes
- [x] T070 [P] [US3] Implement error handling and logging in workers/manim-renderer/src/renderer.py for Manim failures
- [x] T071 [P] [US3] Implement renderVideo in ManimRendererService in workers/ai-manim-worker/src/services/manim.ts to call renderer container
- [x] T072 [US3] Implement uploadToR2 in ManimRendererService in workers/ai-manim-worker/src/services/manim.ts to upload generated video to R2 bucket
- [x] T073 [US3] Implement generatePresignedUrl in ManimRendererService in workers/ai-manim-worker/src/services/manim.ts to create single-access video URL
- [x] T074 [US3] Implement updateJobWithVideo in workers/ai-manim-worker/src/services/session.ts to update status to ready with video_url and expires_at
- [x] T075 [US3] Add status transitions to updating in TelegramHandler: code_validating, rendering, uploading
- [x] T076 [US3] Implement retry logic in TelegramHandler for failed video generation attempts
- [x] T077 [US3] Add logging for video generation flow in workers/ai-manim-worker/src/handlers/telegram.ts (status updates, render duration, video size)
- [x] T078 [US3] Configure R2 lifecycle rules via Cloudflare dashboard for 24-hour auto-delete fallback

### User Story 3 Integration Relays (WORKER↔RENDERER)

- [x] T079-RELAY [US3] Implement Worker→Renderer relay in workers/ai-manim-worker/src/services/manim.ts submitting render requests via HTTP client
- [x] T080-RELAY [US3] Implement Renderer→Worker relay in workers/manim-renderer/src/callback.py sending completion webhooks with job_id
- [x] T081-RELAY [US3] Implement Worker→R2 relay in workers/ai-manim-worker/src/services/manim.ts uploading video with metadata tags
- [x] T082-RELAY [US3] Implement Renderer→R2 direct upload in workers/manim-renderer/src/renderer.py to avoid worker bottleneck
- [x] T083-RELAY [US3] Implement Worker→KV relay in workers/ai-manim-worker/src/services/session.ts updating job status on renderer callbacks
- [x] T084-RELAY [US3] Implement status sync relay between Worker and Renderer in workers/ai-manim-worker/src/utils/renderer-client.ts with polling fallback

**Checkpoint**: All three user stories should now be independently functional. Complete flow: problem → AI → video generation.

---

## Phase 6: User Story 4 - Video Delivery (Priority: P2)

**Goal**: Deliver generated videos to users through their interface with immediate deletion after access

**Independent Test**: Users receive videos through their preferred communication channel within acceptable time limits; videos are deleted immediately after access

### Implementation for User Story 4

- [x] T085 [P] [US4] Implement sendVideoDelivery in workers/ai-manim-worker/src/handlers/telegram.ts to send video URL to user when job status=ready
- [x] T086 [P] [US4] Implement createVideoLink in workers/ai-manim-worker/src/services/manim.ts to generate accessible web link for video
- [x] T087 [P] [US4] Implement trackVideoAccess in workers/ai-manim-worker/src/services/session.ts to update job status to delivered when accessed
- [x] T088 [P] [US4] Implement deleteVideoFromR2 in workers/manim-renderer/src/cleanup.py to immediately delete video after access
- [x] T089 [P] [US4] Implement updateSessionHistory in workers/ai-manim-worker/src/services/session.ts to add VideoMetadata to video_history array
- [x] T090 [US4] Implement GET /api/v1/video/{job_id}/access endpoint in workers/ai-manim-worker/src/handlers/video.ts that redirects to presigned URL
- [x] T091 [US4] Implement immediate video deletion callback after successful delivery in workers/ai-manim-worker/src/handlers/video.ts
- [x] T092 [US4] Implement GET /api/v1/jobs/{job_id} endpoint in workers/ai-manim-worker/src/handlers/video.ts for job status checking
- [x] T093 [US4] Implement GET /api/v1/jobs endpoint in workers/ai-manim-worker/src/handlers/video.ts for listing session jobs
- [x] T094 [US4] Add expiration handling in workers/ai-manim-worker/src/handlers/video.ts to return 410 for already accessed videos
- [x] T095 [US4] Add mobile-optimized formatting in workers/ai-manim-worker/src/handlers/telegram.ts for video delivery messages
- [x] T096 [US4] Add logging for video delivery flow in workers/ai-manim-worker/src/handlers/video.ts (delivery confirmation, access tracking, deletion)

### User Story 4 Integration Relays

- [x] T097-RELAY [US4] Implement KV→Telegram relay in workers/ai-manim-worker/src/handlers/telegram.ts delivering ready jobs to users
- [x] T098-RELAY [US4] Implement VideoAccess→KV relay in workers/ai-manim-worker/src/handlers/video.ts updating delivered status on access
- [x] T099-RELAY [US4] Implement KV→R2 deletion relay in workers/ai-manim-worker/src/handlers/video.ts triggering immediate delete on delivery
- [x] T100-RELAY [US4] Implement R2→SessionHistory relay in workers/ai-manim-worker/src/services/session.ts adding metadata after access
- [x] T101-RELAY [US4] Implement cross-component access tracking relay in workers/ai-manim-worker/src/services/session.ts correlating job_id across KV, R2, and Telegram

**Checkpoint**: User Story 4 complete. Full end-to-end flow works: problem → AI → video → delivery → deletion.

---

## Phase 7: Core Infrastructure for Multi-Platform Support

**Purpose**: Extend existing system to support multiple messaging platforms and direct code submission

### Extend Data Models

- [x] T102 [P] Add Platform enum to types in workers/ai-manim-worker/src/types/index.ts
- [x] T103 [P] Update UserSession interface to support multiple platform IDs in workers/ai-manim-worker/src/types/index.ts
- [x] T104 [P] Extend ProcessingJob interface to include submission_type in workers/ai-manim-worker/src/types/index.ts
- [x] T105 [P] Add WhatsAppMessage interface to types in workers/ai-manim-worker/src/types/index.ts

### Update Session Service

- [x] T106 Update SessionService to handle multiple platform IDs in workers/ai-manim-worker/src/services/session.ts
- [x] T107 Add platform-specific session creation in workers/ai-manim-worker/src/services/session.ts
- [x] T108 Implement cross-platform session lookup in workers/ai-manim-worker/src/services/session.ts

---

## Phase 8: User Story 5 - Direct Code Submission (Priority: P1) 🎯 Enhanced MVP

**Goal**: Allow users to submit Manim code directly and receive video generation without AI processing

**Independent Test**: Users can submit valid Manim code and receive rendered video; invalid code receives helpful error messages

### Implementation for Direct Code Submission

- [x] T109 [P] [US5] Add submission_type validation in workers/ai-manim-worker/src/handlers/code.ts
- [x] T110 [P] [US5] Implement parseCodeSubmission in workers/ai-manim-worker/src/handlers/code.ts to extract Manim code and options
- [x] T111 [P] [US5] Implement validateManimCode in workers/ai-manim-worker/src/handlers/code.ts with syntax and security checks
- [x] T112 [P] [US5] Implement createCodeJob in workers/ai-manim-worker/src/services/session.ts for direct code submissions
- [x] T113 [US5] Implement processCodeSubmission in workers/ai-manim-worker/src/handlers/code.ts to orchestrate direct rendering
- [x] T114 [US5] Implement sendCodeConfirmation in workers/ai-manim-worker/src/handlers/code.ts for submission acknowledgment
- [x] T115 [US5] Implement sendCodeError in workers/ai-manim-worker/src/handlers/code.ts for validation failures
- [x] T116 [US5] Add code submission bypass in TelegramHandler to detect Manim code patterns in workers/ai-manim-worker/src/handlers/telegram.ts
- [x] T117 [US5] Add logging for code submission flow in workers/ai-manim-worker/src/handlers/code.ts

### Direct Code API Endpoints

- [x] T118 [P] [US5] Add POST /api/v1/code endpoint in workers/ai-manim-worker/src/index.ts for direct code submission
- [x] T119 [P] [US5] Add GET /api/v1/code/validate endpoint in workers/ai-manim-worker/src/index.ts for code syntax checking
- [x] T120 [P] [US5] Update video access endpoint to handle code jobs in workers/ai-manim-worker/src/handlers/video.ts

### Direct Code Integration Relays

- [x] T121-RELAY [US5] Implement Code→Renderer relay in workers/ai-manim-worker/src/handlers/code.ts bypassing AI generation
- [x] T122-RELAY [US5] Implement Code→KV relay in workers/ai-manim-worker/src/services/session.ts with submission_type metadata
- [x] T123-RELAY [US5] Implement Renderer→Status relay for code jobs in workers/ai-manim-worker/src/services/manim.ts

---

## Phase 9: User Story 6 - WhatsApp Integration (Priority: P1) 🎯 Platform Expansion

**Goal**: Enable WhatsApp users to submit problems and code for video generation

**Independent Test**: WhatsApp users can submit problems/code and receive video links through WhatsApp messages

### WhatsApp Webhook Infrastructure

- [x] T124 [P] [US6] Implement validateWhatsAppWebhook in workers/ai-manim-worker/src/handlers/whatsapp.ts to verify webhook signatures
- [x] T125 [P] [US6] Implement parseWhatsAppMessage in workers/ai-manim-worker/src/handlers/whatsapp.ts to extract user messages
- [x] T126 [P] [US6] Implement WhatsAppMessageHandler base structure in workers/ai-manim-worker/src/handlers/whatsapp.ts
- [x] T127 [US6] Add POST /webhook/whatsapp route in workers/ai-manim-worker/src/index.ts with webhook validation

### WhatsApp Message Processing

- [x] T128 [P] [US6] Implement detectMessageType in workers/ai-manim-worker/src/handlers/whatsapp.ts for text vs code detection
- [x] T129 [US6] Implement handleWhatsAppProblem in workers/ai-manim-worker/src/handlers/whatsapp.ts for problem submissions
- [x] T130 [US6] Implement handleWhatsAppCode in workers/ai-manim-worker/src/handlers/whatsapp.ts for code submissions
- [x] T131 [US6] Implement sendWhatsAppMessage in workers/ai-manim-worker/src/handlers/whatsapp.ts for user responses
- [x] T132 [US6] Implement sendWhatsAppVideoLink in workers/ai-manim-worker/src/handlers/whatsapp.ts for video delivery
- [x] T133 [US6] Implement sendWhatsAppError in workers/ai-manim-worker/src/handlers/whatsapp.ts for error handling

### WhatsApp Session Management

- [x] T134 [P] [US6] Add WhatsApp session handling in workers/ai-manim-worker/src/services/session.ts with WhatsApp phone numbers
- [x] T135 [US6] Implement createWhatsAppSession in workers/ai-manim-worker/src/services/session.ts for new users
- [x] T136 [US6] Add WhatsApp session auto-extend logic in workers/ai-manim-worker/src/services/session.ts

### WhatsApp External Integration

- [x] T137 [P] [US6] Implement WhatsAppApiClient in workers/ai-manim-worker/src/services/whatsapp.ts for external API calls
- [x] T138 [P] [US6] Add WhatsApp media upload handling in workers/ai-manim-worker/src/services/whatsapp.ts
- [x] T139 [P] [US6] Implement WhatsApp webhook verification in workers/ai-manim-worker/src/services/whatsapp.ts
- [x] T140 [US6] Add WhatsApp rate limiting in workers/ai-manim-worker/src/middleware/rate-limit.ts

### WhatsApp Integration Relays

- [x] T141-RELAY [US6] Implement WhatsApp→Session relay in workers/ai-manim-worker/src/handlers/whatsapp.ts with phone number mapping
- [x] T142-RELAY [US6] Implement WhatsApp→AI/Code relay in workers/ai-manim-worker/src/handlers/whatsapp.ts routing to appropriate handler
- [x] T143-RELAY [US6] Implement Session→WhatsApp relay in workers/ai-manim-worker/src/handlers/whatsapp.ts for message delivery

---

## Phase 10: Platform Abstraction & Cross-Platform Features

**Purpose**: Unify Telegram and WhatsApp handling with shared components

### Abstract Message Handler

- [x] T144 [P] Create BaseMessageHandler abstract class in workers/ai-manim-worker/src/handlers/base.ts
- [x] T145 [P] Refactor TelegramHandler to extend BaseMessageHandler in workers/ai-manim-worker/src/handlers/telegram.ts
- [x] T146 [P] Refactor WhatsAppHandler to extend BaseMessageHandler in workers/ai-manim-worker/src/handlers/whatsapp.ts
- [x] T147 [P] Create MessageRouter in workers/ai-manim-worker/src/handlers/router.ts to route by platform

### Cross-Platform Features

- [x] T148 [P] Add platform detection in workers/ai-manim-worker/src/utils/platform-detector.ts
- [x] T149 [P] Implement unified message formatting in workers/ai-manim-worker/src/utils/message-formatter.ts
- [x] T150 [P] Add cross-platform job status updates in workers/ai-manim-worker/src/services/session.ts
- [x] T151 [P] Create unified error handling in workers/ai-manim-worker/src/utils/error-handler.ts

### Enhanced Web Dashboard

- [ ] T152 [P] Add platform selector to web dashboard in workers/ai-manim-worker/public/dashboard.html
- [ ] T153 [P] Implement direct code submission UI in workers/ai-manim-worker/public/dashboard.html with code editor
- [ ] T154 [P] Add WhatsApp connection instructions in workers/ai-manim-worker/public/dashboard.html
- [ ] T155 [P] Update dashboard JavaScript for direct code submission in workers/ai-manim-worker/public/scripts/dashboard.js
- [ ] T156 [P] Add code editor with syntax highlighting in workers/ai-manim-worker/public/scripts/dashboard.js

---

## Phase 11: Testing, Documentation & Polish

**Purpose**: Ensure quality and documentation for new features

### Testing

- [x] T157 [P] Create unit tests for direct code submission in workers/ai-manim-worker/tests/unit/code.test.ts
- [x] T158 [P] Create unit tests for WhatsApp handler in workers/ai-manim-worker/tests/unit/whatsapp.test.ts
- [ ] T159 [P] Create integration tests for WhatsApp webhook in workers/ai-manim-worker/tests/integration/whatsapp.test.ts
- [ ] T160 [P] Create integration tests for direct code API in workers/ai-manim-worker/tests/integration/code.test.ts

### Documentation

- [ ] T161 [P] Update API documentation for code endpoints in workers/ai-manim-worker/contracts/openapi.yaml
- [ ] T162 [P] Add WhatsApp integration guide in docs/whatsapp-integration.md
- [ ] T163 [P] Update quickstart guide for multi-platform support in docs/quickstart.md
- [ ] T164 [P] Add direct code submission examples in docs/code-submission.md

### Configuration & Deployment

- [x] T165 [P] Update environment variables for WhatsApp in workers/ai-manim-worker/.env.example
- [ ] T166 [P] Add WhatsApp configuration to wrangler.toml in workers/ai-manim-worker/wrangler.toml
- [ ] T167 [P] Update deployment scripts for WhatsApp webhook setup in workers/ai-manim-worker/scripts/deploy.sh

---

## Phase 12: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T168 [P] Update docs/manim-worker.md with deployment instructions and architecture overview
- [x] T169 [P] Create .env.example file at workers/ai-manim-worker/.env.example with all required environment variables documented
- [x] T170 [P] Update AGENTS.md with AI provider configuration details from AGENTS.md
- [x] T171 Implement rate limiting in workers/ai-manim-worker/src/middleware/rate-limit.ts for 10 requests per minute per user
- [x] T172 Implement request timeout handling in workers/ai-manim-worker/src/utils/timeout.ts for 5-minute max processing
- [x] T173 Add CORS configuration in workers/ai-manim-worker/src/index.ts for web dashboard access
- [x] T174 Implement error aggregation in workers/ai-manim-worker/src/utils/errors.ts for better user-facing messages
- [x] T175 Add structured metrics logging in workers/ai-manim-worker/src/utils/metrics.ts for Cloudflare dashboard
- [x] T176 Update wrangler.toml with production environment settings
- [x] T177 Create deployment scripts in workers/ai-manim-worker/scripts/ for worker and renderer deployment
- [x] T178 Run quickstart.md validation to ensure all setup steps are documented correctly

### Cross-Component Integration Relays

- [x] T179-RELAY Implement end-to-end data flow verification in workers/ai-manim-worker/tests/integration/e2e-flow.test.ts
- [x] T180-RELAY Create relay health checks in workers/ai-manim-worker/src/health/relays.ts verifying all component connections
- [x] T181-RELAY Implement distributed tracing in workers/ai-manim-worker/src/utils/tracing.ts tracking requests across Telegram→Worker→AI→Renderer→R2
- [x] T182-RELAY Add circuit breakers in workers/ai-manim-worker/src/utils/circuit-breaker.ts for Worker→Renderer communication failures

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
                              ┌──────────────┘
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
| Code→Renderer | T121-RELAY | CodeHandler→Renderer | Bypass AI generation |
| WhatsApp→Session | T141-RELAY | WhatsAppHandler→SessionService | Phone number mapping |

### Data Flow Across Relays

```
[User submits problem/code]
         │
         ▼
Telegram/WhatsApp → Worker webhook (T036/T127)
         │
         ▼
Session creation/retrieval (T038/T135) → KV write (T048-RELAY)
         │
         ▼
AI generation OR direct code (T050-T055 OR T109-T113) → KV save (T064-RELAY/T122-RELAY)
         │
         ▼
Render request (T079-RELAY/T121-RELAY) → Renderer HTTP
         │
         ▼
Renderer executes (T067-T069) → R2 upload (T082-RELAY)
         │
         ▼
Renderer callback (T080-RELAY) → Worker update (T083-RELAY) → KV
         │
         ▼
Telegram/WhatsApp delivery (T085/T132) → Video URL sent
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
- **User Stories 1-4 (Phases 3-6)**: All depend on Foundational phase completion
  - US1 (Problem Submission) can start after Foundational - NO dependencies on other stories
  - US2 (AI Problem Solving) can start after Foundational - can work in parallel with US1
  - US3 (Video Generation) can start after US2 (needs AI code) or US5 (direct code)
  - US4 (Video Delivery) can start after US3 (needs generated video)
- **Multi-Platform Core (Phase 7)**: Depends on base functionality completion
- **User Stories 5-6 (Phases 8-9)**: Depend on Multi-Platform Core
- **Platform Abstraction (Phase 10)**: Depends on both Direct Code and WhatsApp implementation
- **Testing & Polish (Phases 11-12)**: Depends on all feature implementation

### User Story Dependencies

```
US1 (Telegram Problem) ──────┐
                              ├──> All independent after Foundational
US2 (AI Generation) ────────────┘
         │
         └──> US3 (Video Generation) - needs AI code from US2
                 │
                 └──> US4 (Video Delivery) - needs video from US3

US5 (Direct Code) ──────────┐
                             └──> US3 (can also use direct code)
US6 (WhatsApp) ──────────────┘

US5 (Direct Code) ──────┐
US6 (WhatsApp) ───────────┼──> US10 (Platform Abstraction)
US1 (Telegram) ──────────┘
```

### Parallel Opportunities

**Setup Phase (Phase 1)**:
- T001, T002, T003, T004, T005 can run in parallel
- T006, T007, T008 can run in parallel

**Foundational Phase (Phase 2)**:
- T009-T014 (all type definitions) can run in parallel
- T019-T021 (handler structures) can run in parallel

**Multi-Platform Core (Phase 7)**:
- T102, T103, T104, T105 can run in parallel
- T106, T107, T108 can run in parallel

**Direct Code Submission (Phase 8)**:
- T109, T110, T111, T112 can run in parallel
- T118, T119, T120 can run in parallel

**WhatsApp Integration (Phase 9)**:
- T124, T125, T126, T127 can run in parallel
- T128, T129, T130, T131, T132, T133 can run in parallel
- T137, T138, T139, T140 can run in parallel

**Platform Abstraction (Phase 10)**:
- T144, T145, T146, T147 can run in parallel
- T148, T149, T150, T151 can run in parallel

**Testing & Polish (Phases 11-12)**:
- T157, T158, T159, T160 can run in parallel
- T161, T162, T163, T164 can run in parallel

---

## Implementation Strategy

### MVP Enhancement First (Direct Code + WhatsApp)

1. Complete Phase 1-6: Base functionality (ALREADY DONE ✅)
2. Complete Phase 7: Multi-Platform Core Infrastructure
3. Complete Phase 8: Direct Code Submission
4. **STOP and VALIDATE**: Test direct code submission independently
   - Submit Manim code via WhatsApp/web
   - Receive rendered video directly
   - Verify error handling for invalid code
5. Complete Phase 9: WhatsApp Integration
6. **STOP and VALIDATE**: Test WhatsApp problem/code submission
7. Complete Phase 10: Platform Abstraction
8. Complete Phases 11-12: Testing & Polish
9. Deploy enhanced system

### Incremental Platform Addition

1. Complete Setup + Foundational + Base US (Phases 1-6) → Foundation ready ✅
2. Add Multi-Platform Core (Phase 7) → Multi-platform foundation ready
3. Add Direct Code Submission (Phase 8) → Test independently → Deploy/Demo
4. Add WhatsApp Integration (Phase 9) → Test independently → Deploy/Demo
5. Add Platform Abstraction (Phase 10) → Unified system → Deploy/Demo
6. Each addition adds value without breaking existing features

### Parallel Team Strategy

With multiple developers:

1. Team completes Multi-Platform Core together (Phase 7)
2. Once Core Infrastructure is done:
   - Developer A: Direct Code Submission (Phase 8)
   - Developer B: WhatsApp Integration (Phase 9)
   - Developer C: Platform Abstraction (Phase 10) (after A&B)
3. Features complete and integrate independently

---

## Key Enhancements Summary

### New Capabilities Added

1. **Direct Manim Code Submission**
   - Users bypass AI generation entirely
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

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **RELAY tasks create strong connections** between: Telegram ↔ Worker ↔ AI ↔ Renderer ↔ R2
- Worker and renderer are separate projects that must be coordinated via explicit relay points
- Mock renderer should be used for development to avoid Docker dependencies
- R2 lifecycle rules provide fallback for video deletion; immediate deletion via code is primary
- All AI providers have free tiers; configure at least one for functionality
- Session TTL is 7 days with auto-extend; job TTL varies by status
- **Strong relay architecture ensures** data flow visibility, error propagation, and state synchronization across all components

---

**Total Task Count: 182**
- Setup & Foundation: 45 tasks (24 ✅ completed, 21 ❌ pending)
- Base User Stories (US1-US4): 66 tasks (66 ✅ completed)
- Enhancement Infrastructure: 8 tasks (8 ❌ pending)
- Direct Code Submission (US5): 15 tasks (15 ❌ pending)
- WhatsApp Integration (US6): 20 tasks (20 ❌ pending)
- Platform Abstraction: 13 tasks (13 ❌ pending)
- Testing & Polish: 15 tasks (15 ❌ pending)

**Independent Test Criteria Per Story**:
- US5: Direct code submission → video rendering or clear error
- US6: WhatsApp users → submit problems/code → receive videos
- Base US1-US4: Already validated and working ✅