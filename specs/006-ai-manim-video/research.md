# Research: AI Manim Video Generator

**Feature**: 006-ai-manim-video  
**Date**: January 15, 2026  
**Purpose**: Resolve technical unknowns for implementation

---

## R1: Cloudflare Workers AI for Python Code Generation

**Question**: Can Cloudflare Workers AI generate valid Manim Python code?

### Findings

**Cloudflare Workers AI Capabilities**:
- Supports Llama 2, Mistral, and other open models
- Text generation endpoint can produce Python code
- No native Python execution, only text output

**Feasibility Assessment**: ✅ VIABLE

**Approach**:
1. Use `@cf/meta/llama-2-7b-chat-int8` for code generation
2. Craft prompt with Manim-specific instructions
3. Validate generated code syntax before rendering

**Prompt Strategy**:
```python
# System prompt for Manim code generation
SYSTEM_PROMPT = """
You are a Manim animation expert. Generate Python code using Manim library.
Requirements:
1. Use Manim v0.18+ syntax (Scene, Tex, MathTex, etc.)
2. Include only the animation code, no explanations
3. Output valid Python that can be executed standalone
4. Keep animations under 30 seconds
5. Use appropriate colors and styling
"""
```

**Alternative**: If Cloudflare AI struggles, use Groq API (free tier available).

---

## R2: Manim Rendering in Containers

**Question**: Best practices for Docker-based Manim rendering?

### Findings

**Manim Dependencies**:
- Python 3.11+
- LaTeX (TexLive)
- FFmpeg
- SoX (for audio)
- Cairo/Pango (for SVGs)

**Docker Optimization Strategies**:

| Strategy | Benefit |
|----------|---------|
| Multi-stage build | Smaller final image (~2GB vs 8GB) |
| Pre-installed LaTeX | Faster startup (skip ~2 min install) |
| Resource limits | Prevent runaway renders (memory caps) |
| Timeout enforcement | Max 5-minute render per video |
| Ephemeral containers | One render per container, auto-delete |

**Recommended Base Image**:
```dockerfile
# Multi-stage build for Manim
FROM manimorg/manim:0.18.1 as builder

# Or lighter alternative
FROM python:3.11-slim
RUN apt-get update && apt-get install -y \
    texlive-latex-extra \
    ffmpeg \
    sox \
    libcairo2 \
    libpango1.0-0
RUN pip install manim
```

**Resource Limits**:
- Memory: 2GB max
- CPU: 2 cores
- Timeout: 300 seconds (5 minutes)
- Output: MP4, max 50MB

---

## R3: Telegram Webhook Security

**Question**: How to validate Telegram webhooks and prevent spoofing?

### Findings

**Telegram Webhook Security**:

1. **Token Validation**: Telegram sends bot token in `X-Telegram-Bot-Api-Secret-Token` header
   - Generate random secret during worker deployment
   - Validate on every request

2. **IP Allowlist** (Optional): Telegram uses fixed IPs
   - 149.154.160.0/20 range
   - Validate source IP if paranoid

3. **Signature Verification** (Not needed for webhooks):
   - Only needed for local bot polling
   - Webhooks are already authenticated via bot token

**Implementation**:
```typescript
// Validate Telegram webhook
export async function handleTelegramWebhook(request: Request): Promise<Response> {
  const secretToken = request.headers.get('X-Telegram-Bot-Api-Secret-Token');
  
  if (secretToken !== env.TELEGRAM_SECRET) {
    return new Response('Unauthorized', { status: 401 });
  }
  
  const update = await request.json();
  // Process update...
}
```

---

## R4: Cloudflare R2 + Immediate Delete

**Question**: Workflow for single-access video URLs with immediate deletion?

### Findings

**R2 Single-Access Pattern**:

1. **Upload with Expiration**:
   ```typescript
   // Upload to R2 with auto-delete
   await r2.upload(videoKey, videoData, {
     onlyIf: { etagMatch: null },  // No conditions
     custom: {
       'x-r2-delete-after': '3600',  // Delete after 1 hour
     }
   });
   ```

2. **Presigned URL Generation**:
   ```typescript
   // Generate single-use URL
   const url = await r2.createPresignedUrl({
     method: 'GET',
     key: videoKey,
     expiresIn: 300,  // 5 minute URL
   });
   ```

3. **Delete After Access**:
   - Use R2 lifecycle rules for automatic deletion
   - Or delete immediately after confirmed download

**Recommended Workflow**:
1. Upload video to R2
2. Generate presigned URL (5-minute expiration)
3. Send URL to user via Telegram
4. User accesses video (one-time or 5-min window)
5. R2 lifecycle rule deletes after 24 hours (fallback)
6. Session metadata tracks "delivered" status

---

## R5: AI Fallback Chain Implementation

**Question**: How to implement seamless provider switching?

### Findings

**Fallback Strategy**:

| Provider | Priority | Free Tier | Model |
|----------|----------|-----------|-------|
| Cloudflare Workers AI | 1 | Unlimited | Llama 2 7B |
| Groq | 2 | Free tier | Llama 2 70B |
| HuggingFace | 3 | Free tier | CodeLlama |

**Implementation Pattern**:

```typescript
// Fallback chain service
class AIFallbackService {
  private providers: AIProvider[] = [
    new CloudflareAI(),
    new GroqAI(),
    new HuggingFaceAI()
  ];
  
  async generateManimCode(prompt: string): Promise<string> {
    const errors: Error[] = [];
    
    for (const provider of this.providers) {
      try {
        const code = await provider.generate(prompt);
        await this.logSuccess(provider.name);
        return code;
      } catch (error) {
        errors.push(error);
        await this.logFailure(provider.name, error);
        continue;
      }
    }
    
    throw new AggregateError(errors, 'All AI providers failed');
  }
}
```

**Health Monitoring**:
- Track success rate per provider
- Auto-degrade priority if error rate > 10%
- Log all failures for debugging

---

## R6: WhatsApp Integration Patterns

**Question**: How to integrate WhatsApp for programmatic messaging and direct code submission?

### Findings

**WhatsApp API Options**:

1. **WhatsApp Business API** (Recommended):
   - Official enterprise solution
   - HTTP API for sending/receiving messages
   - Webhook support for incoming messages
   - Template messaging for marketing
   - Phone number verification required

2. **WhatsApp Cloud API**:
   - Meta-hosted version of Business API
   - Easier setup for developers
   - Similar capabilities but cloud-managed

**Integration Approach**:
- Use WhatsApp Business API with HTTP webhooks
- Implement message parsing for direct code detection
- Support both text and media messages
- Handle authentication via API keys

**Security Considerations**:
- API key management (rotate regularly)
- Webhook signature validation
- Rate limiting (WhatsApp has strict limits)
- Content moderation for user-generated code

**Implementation Pattern**:
```typescript
class WhatsAppService {
  async sendMessage(phoneNumber: string, message: string): Promise<void> {
    // Send via WhatsApp Business API
  }

  async handleWebhook(payload: WhatsAppWebhook): Promise<void> {
    // Process incoming messages
    // Detect Manim code patterns
    // Route to appropriate handler
  }
}
```

**Cost Structure**:
- Free tier available for low-volume testing
- Pay-per-message for production
- Phone number leasing required

**Decision**: Use WhatsApp Business API with webhook integration for reliable, programmatic messaging with direct code submission support.

---

## Decisions Summary

| Topic | Decision | Rationale |
|-------|----------|-----------|
| AI Provider | Cloudflare Workers AI + fallback | Free tier, edge computing, reliable |
| Manim Rendering | Docker container with limits | Security, reproducibility, resource caps |
| Telegram Security | Secret token validation | Simple, effective, industry standard |
| Video Storage | R2 with presigned URLs + auto-delete | Zero retention, single access |
| Fallback Chain | Priority-based with health tracking | Reliability without cost |

## Alternatives Considered

| Alternative | Why Rejected |
|-------------|--------------|
| Local Manim server | No free compute; maintenance burden |
| Paid API (Replicate) | Violates "Free-Only AI" constitution |
| Real-time video streaming | Storage/complexity overhead |
| User-provided rendering | Poor UX, unreliable |

## Open Questions (for planning phase)

1. **R2 costs**: Need to estimate storage/bandwidth for 100 videos/day
2. **Container scaling**: Auto-scale or fixed pool? (Cost vs latency trade-off)
3. **Retry strategy**: How many times to retry failed renders?
4. **Rate limits**: Per-user limits to prevent abuse

---

**Research completed**: January 15, 2026  
**Next**: Phase 1 - Data model and API contracts
