# Implementation Plan: AI Manim Video Generator

**Branch**: `006-ai-manim-video` | **Date**: January 15, 2026 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification for AI-powered mathematical video generation

## Summary

Build a Cloudflare Workers-based video generation service where users submit mathematical problems via Telegram, an AI generates Manim animation code, and the resulting video is delivered via web dashboard link with immediate deletion after access. Uses Cloudflare Workers AI for code generation with fallback providers, anonymous session tracking, and zero-persistent storage architecture.

## Technical Context

**Language/Version**: TypeScript/JavaScript (Cloudflare Workers) | Python 3.11 (Manim execution)  
**Primary Dependencies**: Cloudflare Workers, Cloudflare Workers AI, Telegram Bot API, Manim library  
**Storage**: Cloudflare KV for session metadata | R2 for temporary video storage (immediate delete)  
**Testing**: Jest (unit), Integration tests via wrangler | Python pytest for Manim validation  
**Target Platform**: Cloudflare Workers (edge), Python containers for rendering  
**Project Type**: Worker-based service with external rendering pipeline  
**Performance Goals**: 5-minute video generation, 10 concurrent requests, <50MB video output  
**Constraints**: Zero video retention, anonymous users only, no external paid APIs  
**Scale/Scope**: Initial: 100 users/day, expandable to 1000 concurrent

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Free-Only AI** | ✅ PASS | Cloudflare Workers AI is free; no paid APIs |
| **II. Privacy-First** | ✅ PASS | Anonymous sessions, no data retention, immediate video deletion |
| **III. Test-First (NON-NEGOTIABLE)** | ⚠️ NEEDS VERIFICATION | Tests must precede implementation |
| **IV. Integration Testing** | ✅ REQUIRED | Telegram webhook, AI fallback chain, video delivery |
| **V. Observability & Simplicity** | ✅ REQUIRED | Structured logging, Cloudflare dashboard metrics |

## Constitution Compliance Notes

- Cloudflare Workers AI is the primary free provider
- All user data is session-based with 7-day expiration
- Videos deleted immediately after delivery - zero persistent storage
- Fallback to alternative providers ensures reliability without paid services

## Project Structure

### Documentation (this feature)

```text
specs/006-ai-manim-video/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (pending)
├── data-model.md        # Phase 1 output (pending)
├── quickstart.md        # Phase 1 output (pending)
├── contracts/           # Phase 1 output (pending)
│   └── openapi.yaml
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (Cloudflare Workers + Python)

```text
workers/ai-manim-worker/
├── src/
│   ├── index.ts         # Worker entry point
│   ├── handlers/
│   │   ├── telegram.ts  # Telegram webhook handler
│   │   ├── ai.ts        # AI code generation
│   │   └── video.ts     # Video generation orchestration
│   ├── services/
│   │   ├── session.ts   # Session management
│   │   ├── fallback.ts  # AI provider fallback
│   │   └── manim.ts     # Manim rendering service
│   ├── types/
│   │   └── index.ts     # TypeScript types
│   └── utils/
│       └── logger.ts    # Structured logging
├── tests/
│   ├── unit/
│   │   ├── session.test.ts
│   │   └── fallback.test.ts
│   └── integration/
│       ├── telegram.test.ts
│       └── video.test.ts
├── package.json
├── wrangler.toml
└── tsconfig.json

workers/manim-renderer/
├── src/
│   ├── renderer.py      # Manim execution
│   └── cleanup.py       # Video deletion
├── Dockerfile
└── requirements.txt

docs/
└── manim-worker.md      # Worker documentation
```

**Structure Decision**: Cloudflare Worker for API/coordination + Python container for Manim rendering (Docker-based for security isolation). Session metadata in KV, videos in R2 with immediate deletion.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Docker container for rendering | Manim requires system dependencies; Workers can't run it | Python Workers available but limited; container provides full Manim support |

## Technical Decisions

### AI Provider Strategy

**Decision**: Cloudflare Workers AI primary with fallback chain

**Providers (Priority Order)**:
1. Cloudflare Workers AI (@cf/meta/llama-2-7b-chat)
2. Groq (free tier available)
3. HuggingFace Inference API (free tier)

**Rationale**: All providers have free tiers, ensuring zero cost while maintaining reliability through fallback.

### Video Rendering Pipeline

**Decision**: Containerized Manim execution with Cloudflare R2 storage

**Flow**:
1. Telegram → Worker (problem submission)
2. Worker → AI (code generation)
3. Worker → Container (render video)
4. Container → R2 (temporary storage)
5. Worker → Telegram (web link)
6. User → Web (video access)
7. R2 → Delete (immediate after access)

### Session Management

**Decision**: Cloudflare KV for session data

**Session Schema**:
- session_id: string (7-day TTL)
- telegram_chat_id: string
- created_at: timestamp
- last_activity: timestamp (auto-extend)
- video_history: array (metadata only, no videos)

## Research Requirements (Phase 0)

Based on Technical Context, research these unknowns:

1. **Cloudflare Workers AI for Python code generation** - Can it generate valid Manim code?
2. **Manim rendering in containers** - Best practices for Docker + resource limits
3. **Telegram webhook security** - Validate bot token, prevent spoofing
4. **Cloudflare R2 + immediate delete** - Workflow for single-access URLs
5. **AI fallback chain** - How to implement seamless provider switching

## Design Requirements (Phase 1)

After research, create:

1. **data-model.md** - Session, ProcessingJob, VideoFile entities
2. **contracts/openapi.yaml** - Telegram webhook, video delivery endpoints
3. **quickstart.md** - Development setup guide
4. **Update agent context** - Add Cloudflare Workers to context

## Next Steps

1. **Phase 0**: Run `.specify/scripts/bash/research.sh` to resolve unknowns
2. **Phase 1**: Generate data-model.md, contracts, quickstart.md
3. **Phase 2**: Run `/speckit.tasks` to create implementation tasks
