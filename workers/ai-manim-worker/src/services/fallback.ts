import type { Env } from "../types";
import { createLogger } from "../utils/logger";

interface AIProvider {
  name: string;
  generate(prompt: string): Promise<string>;
  isAvailable(): Promise<boolean>;
}

class CloudflareAIProvider implements AIProvider {
  name = "cloudflare";
  private env: Env;
  private logger: ReturnType<typeof createLogger>;

  constructor(env: Env) {
    this.env = env;
    this.logger = createLogger({
      level: (env.LOG_LEVEL as "debug" | "info" | "warn" | "error") || "info",
      component: "ai-provider",
    });
  }

  async generate(prompt: string): Promise<string> {
    this.logger.info("Generating with Cloudflare AI", { prompt_length: prompt.length });

    const response = await fetch("https://api.cloudflare.com/client/v4/accounts/" + 
      this.env.CLOUDFLARE_ACCOUNT_ID + "/ai/run/@cf/meta/llama-2-7b-chat-int8", {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.env.CLOUDFLARE_API_TOKEN}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        messages: [
          { role: "user", content: prompt }
        ],
        max_tokens: 2048,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      this.logger.error("Cloudflare AI failed", { status: response.status, error });
      throw new Error(`Cloudflare AI error: ${response.status}`);
    }

    const data = await response.json();
    const result = data.result?.response || "";

    this.logger.info("Cloudflare AI generated code", { code_length: result.length });
    return result;
  }

  async isAvailable(): Promise<boolean> {
    return true;
  }
}

class GroqAIProvider implements AIProvider {
  name = "groq";
  private env: Env;
  private logger: ReturnType<typeof createLogger>;

  constructor(env: Env) {
    this.env = env;
    this.logger = createLogger({
      level: (env.LOG_LEVEL as "debug" | "info" | "warn" | "error") || "info",
      component: "ai-provider",
    });
  }

  async generate(prompt: string): Promise<string> {
    this.logger.info("Generating with Groq AI", { prompt_length: prompt.length });

    const response = await fetch("https://api.groq.com/openai/v1/chat/completions", {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.env.GROQ_API_KEY}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        model: "llama2-70b-4096",
        messages: [
          { role: "user", content: prompt }
        ],
        max_tokens: 2048,
        temperature: 0.1,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      this.logger.error("Groq AI failed", { status: response.status, error });
      throw new Error(`Groq AI error: ${response.status}`);
    }

    const data = await response.json();
    const result = data.choices?.[0]?.message?.content || "";

    this.logger.info("Groq AI generated code", { code_length: result.length });
    return result;
  }

  async isAvailable(): Promise<boolean> {
    return !!this.env.GROQ_API_KEY;
  }
}

class HuggingFaceProvider implements AIProvider {
  name = "huggingface";
  private env: Env;
  private logger: ReturnType<typeof createLogger>;

  constructor(env: Env) {
    this.env = env;
    this.logger = createLogger({
      level: (env.LOG_LEVEL as "debug" | "info" | "warn" | "error") || "info",
      component: "ai-provider",
    });
  }

  async generate(prompt: string): Promise<string> {
    this.logger.info("Generating with HuggingFace AI", { prompt_length: prompt.length });

    const response = await fetch("https://api-inference.huggingface.co/models/codellama/CodeLlama-7b-instruct-hf", {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.env.HF_TOKEN}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        inputs: prompt,
        parameters: {
          max_new_tokens: 2048,
          temperature: 0.1,
        },
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      this.logger.error("HuggingFace AI failed", { status: response.status, error });
      throw new Error(`HuggingFace AI error: ${response.status}`);
    }

    const data = await response.json();
    const result = Array.isArray(data) ? data[0]?.generated_text || "" : data.generated_text || "";

    this.logger.info("HuggingFace AI generated code", { code_length: result.length });
    return result;
  }

  async isAvailable(): Promise<boolean> {
    return !!this.env.HF_TOKEN;
  }
}

export class AIFallbackService {
  private env: Env;
  private logger: ReturnType<typeof createLogger>;
  private providers: AIProvider[];

  constructor(env: Env) {
    this.env = env;
    this.logger = createLogger({
      level: (env.LOG_LEVEL as "debug" | "info" | "warn" | "error") || "info",
      component: "ai-fallback",
    });

    this.providers = [
      new CloudflareAIProvider(env),
      new GroqAIProvider(env),
      new HuggingFaceProvider(env),
    ];
  }

  async generateManimCode(problem: string): Promise<string> {
    const systemPrompt = `You are a Manim animation expert. Generate Python code using Manim library.
Requirements:
1. Use Manim v0.18+ syntax (Scene, Tex, MathTex, etc.)
2. Output only the Python code, no markdown, no explanations
3. Code must be valid Python that can be executed standalone
4. Keep animations under 30 seconds
5. Use appropriate colors and styling

Problem to visualize:
${problem}`;

    const errors: Error[] = [];

    for (const provider of this.providers) {
      try {
        if (!(await provider.isAvailable())) {
          this.logger.debug("Provider not available", { provider: provider.name });
          continue;
        }

        this.logger.info("Trying AI provider", { provider: provider.name });
        const code = await provider.generate(systemPrompt);
        
        this.logger.info("AI generation successful", { 
          provider: provider.name, 
          code_length: code.length 
        });

        return code;
      } catch (error) {
        errors.push(error as Error);
        this.logger.warn("AI provider failed", { 
          provider: provider.name, 
          error: (error as Error).message 
        });
        continue;
      }
    }

    const allErrors = errors.map(e => e.message).join("; ");
    this.logger.error("All AI providers failed", { errors: allErrors });
    throw new Error(`All AI providers failed: ${allErrors}`);
  }

  getProvidersStatus(): { name: string; available: boolean }[] {
    return this.providers.map(p => ({
      name: p.name,
      available: true,
    }));
  }
}

export function createAIFallbackService(env: Env): AIFallbackService {
  return new AIFallbackService(env);
}
