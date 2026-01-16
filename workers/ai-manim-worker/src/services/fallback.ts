import type { Env } from "../types";
import { createLogger } from "../utils/logger";

interface AIProvider {
  name: string;
  generate(prompt: string): Promise<string>;
  isAvailable(): Promise<boolean>;
}

class ProviderError extends Error {
  constructor(
    message: string,
    public readonly provider: string,
    public readonly statusCode?: number
  ) {
    super(message);
    this.name = 'ProviderError';
  }
}

class OpenAIProvider implements AIProvider {
  name = "openai";
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
    this.logger.info("Generating with OpenAI", { prompt_length: prompt.length });

    const response = await fetch("https://api.openai.com/v1/chat/completions", {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.env.OPENAI_API_KEY}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        model: "gpt-4o",
        messages: [
          { role: "system", content: "You are a Python code generator for the Manim library." },
          { role: "user", content: prompt }
        ],
        max_tokens: 2048,
        temperature: 0.1,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      this.logger.error("OpenAI failed", { status: response.status, error: errorText });
      throw new ProviderError(`OpenAI error: ${response.status}`, this.name, response.status);
    }

    const data = await response.json() as any;
    const result = data.choices?.[0]?.message?.content || "";

    this.logger.info("OpenAI generated code", { code_length: result.length });
    return result;
  }

  async isAvailable(): Promise<boolean> {
    return !!this.env.OPENAI_API_KEY;
  }
}

class OpenAIProvider implements AIProvider {
  name = "openai";
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
    this.logger.info("Generating with OpenAI", { prompt_length: prompt.length });

    const response = await fetch("https://api.openai.com/v1/chat/completions", {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.env.OPENAI_API_KEY}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        model: "gpt-4o",
        messages: [
          { role: "system", content: "You are a Python code generator for the Manim library." },
          { role: "user", content: prompt }
        ],
        max_tokens: 2048,
        temperature: 0.1,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      this.logger.error("OpenAI failed", { status: response.status, error: errorText });
      throw new Error(`OpenAI error: ${response.status}`);
    }

    const data = await response.json() as any;
    const result = data.choices?.[0]?.message?.content || "";

    this.logger.info("OpenAI generated code", { code_length: result.length });
    return result;
  }

  async isAvailable(): Promise<boolean> {
    return !!this.env.OPENAI_API_KEY;
  }
}

class GeminiProvider implements AIProvider {
  name = "gemini";
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
    this.logger.info("Generating with Gemini", { prompt_length: prompt.length });

    const response = await fetch(`https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-pro:generateContent?key=${this.env.GEMINI_API_KEY}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        contents: [{
          parts: [{ text: prompt }]
        }],
        generationConfig: {
          maxOutputTokens: 2048,
          temperature: 0.1,
        },
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      this.logger.error("Gemini failed", { status: response.status, error: errorText });
      throw new ProviderError(`Gemini error: ${response.status}`, this.name, response.status);
    }

    const data = await response.json();
    const result = data.candidates?.[0]?.content?.parts?.[0]?.text || "";

    this.logger.info("Gemini generated code", { code_length: result.length });
    return result;
  }

  async isAvailable(): Promise<boolean> {
    return !!this.env.GEMINI_API_KEY && this.env.GEMINI_API_KEY !== "your-gemini-api-key";
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
        model: "mixtral-8x7b-32768",
        messages: [
          { role: "user", content: prompt }
        ],
        max_tokens: 2048,
        temperature: 0.1,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      this.logger.error("Groq AI failed", { status: response.status, error: errorText });
      throw new ProviderError(`Groq AI error: ${response.status}`, this.name, response.status);
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
      const errorText = await response.text();
      this.logger.error("HuggingFace AI failed", { status: response.status, error: errorText });
      throw new ProviderError(`HuggingFace AI error: ${response.status}`, this.name, response.status);
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

class DeepSeekProvider implements AIProvider {
  name = "deepseek";
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
    this.logger.info("Generating with DeepSeek", { prompt_length: prompt.length });

    const response = await fetch("https://api.deepseek.com/v1/chat/completions", {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.env.DEEPSEEK_API_KEY}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        model: "deepseek-coder",
        messages: [
          { role: "user", content: prompt }
        ],
        max_tokens: 2048,
        temperature: 0.1,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      this.logger.error("DeepSeek failed", { status: response.status, error: errorText });
      throw new ProviderError(`DeepSeek error: ${response.status}`, this.name, response.status);
    }

    const data = await response.json();
    const result = data.choices?.[0]?.message?.content || "";

    this.logger.info("DeepSeek generated code", { code_length: result.length });
    return result;
  }

  async isAvailable(): Promise<boolean> {
    return !!this.env.DEEPSEEK_API_KEY;
  }
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
      const errorText = await response.text();
      this.logger.error("Cloudflare AI failed", { status: response.status, error: errorText });
      throw new ProviderError(`Cloudflare AI error: ${response.status}`, this.name, response.status);
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

export class AIFallbackService {
  private env: Env;
  private logger: ReturnType<typeof createLogger>;
  private providers: AIProvider[];
  private providerStats: Map<string, { success: number; total: number; lastError?: string }>;

  constructor(env: Env) {
    this.env = env;
    this.logger = createLogger({
      level: (env.LOG_LEVEL as "debug" | "info" | "warn" | "error") || "info",
      component: "ai-fallback",
    });

    this.providers = [
      new OpenAIProvider(env),
      new GeminiProvider(env),
      new DeepSeekProvider(env),
      new GroqAIProvider(env),
      new HuggingFaceProvider(env),
      new CloudflareAIProvider(env),
    ];

    this.providerStats = new Map();
  }

  async generateManimCode(problem: string): Promise<string> {
    const systemPrompt = `You are a Manim animation expert. Generate Python code using Manim library.
Requirements:
1. Use Manim v0.18+ syntax (Scene, Tex, MathTex, etc.)
2. Output only the Python code, no markdown code blocks, no explanations
3. Code must be valid Python that can be executed standalone
4. Keep animations under 30 seconds
5. Use appropriate colors and styling
6. The code should start with 'from manim import *' and include a Scene class

Problem to visualize:
${problem}`;

    for (const provider of this.providers) {
      try {
        const isAvailable = await provider.isAvailable();
        if (!isAvailable) {
          this.logger.debug("Provider not available", { provider: provider.name });
          continue;
        }

        this.logger.info("Trying AI provider", { provider: provider.name });
        const code = await provider.generate(systemPrompt);

        this.logger.info("AI generation successful", {
          provider: provider.name,
          code_length: code.length
        });

        this.recordSuccess(provider.name);
        return code;
      } catch (error) {
        this.recordFailure(provider.name, error as ProviderError);
        this.logger.warn("AI provider failed", {
          provider: provider.name,
          error: (error as Error).message
        });
        continue;
      }
    }

    const allErrors = Array.from(this.providerStats.values())
      .filter(stats => stats.lastError)
      .map(stats => stats.lastError!)
      .join("; ");

    this.logger.error("All AI providers failed", { errors: allErrors });
    throw new Error(`All AI providers failed: ${allErrors}`);
  }

  private recordSuccess(providerName: string): void {
    const stats = this.providerStats.get(providerName) || { success: 0, total: 0 };
    stats.success++;
    stats.total++;
    stats.lastError = undefined;
    this.providerStats.set(providerName, stats);
  }

  private recordFailure(providerName: string, error: ProviderError): void {
    const stats = this.providerStats.get(providerName) || { success: 0, total: 0 };
    stats.total++;
    stats.lastError = error.message;
    this.providerStats.set(providerName, stats);
  }

  getProvidersStatus(): { name: string; available: boolean; successRate?: number }[] {
    return this.providers.map(provider => {
      const stats = this.providerStats.get(provider.name);
      const successRate = stats ? stats.total > 0 ? stats.success / stats.total : undefined : undefined;

      return {
        name: provider.name,
        available: true,
        successRate,
      };
    });
  }
}

export function createAIFallbackService(env: Env): AIFallbackService {
  return new AIFallbackService(env);
}
