# Implementation Plan: AI Manim Video Generator

**Branch**: `006-ai-manim-video` | **Date**: January 17, 2026 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification for AI-powered mathematical video generation

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a Cloudflare Workers-based video generation service where users submit mathematical problems via Telegram/WhatsApp, an AI generates Manim animation code, and the resulting video is delivered via web dashboard link with immediate deletion after access. Uses Cloudflare Workers AI for code generation with fallback providers, anonymous session tracking, and zero-persistent storage architecture.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: TypeScript/JavaScript (Cloudflare Workers) | Python 3.11 (Manim execution)
**Primary Dependencies**: Cloudflare Workers, Cloudflare Workers AI, Telegram/WhatsApp Bot APIs, Manim library
**Storage**: Cloudflare KV for session metadata | R2 for temporary video storage (immediate delete)
**Testing**: Jest (unit), Integration tests via wrangler | Python pytest for Manim validation
**Target Platform**: Cloudflare Workers (edge), Python containers for rendering
**Project Type**: Worker-based service with external rendering pipeline
**Performance Goals**: Flexible scaling (no strict limits defined)
**Constraints**: Zero video retention, anonymous users only, no external paid APIs, WCAG 2.1 AAA accessibility
**Scale/Scope**: Initial: 100 users/day, expandable to 1000 concurrent

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Free-Only AI** | ✅ PASS | Cloudflare Workers AI is free; fallback providers are free-tier available |
| **II. Privacy-First** | ✅ PASS | Anonymous sessions, no data retention, immediate video deletion |
| **III. Test-First (NON-NEGOTIABLE)** | ⚠️ NEEDS VERIFICATION | Tests must precede implementation |
| **IV. Integration Testing** | ✅ REQUIRED | Telegram/WhatsApp webhook, AI fallback chain, video delivery |
| **V. Observability & Simplicity** | ✅ REQUIRED | Structured logging, Cloudflare dashboard metrics |

**Quality Standards Compliance:**
- **Performance**: Flexible scaling approach (constitution specifies <5s for API calls)
- **Security**: Minimal security with basic input validation (constitution requires rate limiting, content filtering)
- **Compliance**: GDPR/COPPA compliance for European users, open source licensing

**Phase 1 Complete**: Data model, API contracts, quickstart guide, and agent context updated.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (Cloudflare Workers + Python)

```text
workers/ai-manim-worker/
├── src/
│   ├── index.ts         # Worker entry point
│   ├── handlers/
│   │   ├── telegram.ts  # Telegram webhook handler
│   │   ├── whatsapp.ts  # WhatsApp webhook handler
│   │   ├── code.ts      # Direct code submission handler
│   │   ├── video.ts     # Video delivery handler
│   │   └── debug.ts     # Health checks
│   ├── services/
│   │   ├── session.ts   # Session management
│   │   ├── fallback.ts  # AI provider fallback
│   │   ├── manim.ts     # Manim rendering service
│   │   ├── whatsapp.ts  # WhatsApp API client
│   │   └── telegram.ts  # Telegram API client
│   ├── types/
│   │   └── index.ts     # TypeScript type definitions
│   ├── utils/
│   │   ├── logger.ts    # Structured logging
│   │   ├── error-handler.ts # Error handling utilities
│   │   └── platform-detector.ts # Platform detection
│   └── middleware/
│       └── rate-limit.ts # Rate limiting
├── tests/
│   ├── unit/
│   │   ├── session.test.ts
│   │   └── fallback.test.ts
│   └── integration/
│       ├── telegram.test.ts
│       ├── whatsapp.test.ts
│       └── video.test.ts
├── package.json
├── wrangler.toml
├── tsconfig.json
└── .eslintrc.js

workers/manim-renderer/
├── src/
│   ├── renderer.py      # Manim execution
│   ├── server.py        # HTTP server for worker communication
│   ├── cleanup.py       # Video deletion utilities
│   ├── worker-client.py # Worker communication
│   └── logger.py        # Shared logging
├── Dockerfile
├── requirements.txt
└── tests/
    └── test_renderer.py
```

**Structure Decision**: Cloudflare Worker for API/coordination + Python container for Manim rendering (Docker-based for security isolation). Session metadata in KV, videos in R2 with immediate deletion.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
