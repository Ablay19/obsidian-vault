import type { Env, TelegramUpdate, ProcessingJob, ProcessingStatus } from "./types";
import { createLogger } from "./utils/logger";

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    const logger = createLogger({ level: (env.LOG_LEVEL as "debug" | "info" | "warn" | "error") || "info", component: "ai-manim-worker" });

    logger.info("Incoming request", {
      method: request.method,
      url: url.pathname,
    });

    try {
      if (url.pathname === "/health") {
        return handleHealth();
      }

      if (url.pathname === "/webhook/telegram" && request.method === "POST") {
        return handleTelegramWebhook(request, env, logger);
      }

      if (url.pathname.startsWith("/api/v1/")) {
        return handleAPI(request, env, url, logger);
      }

      return new Response("Not Found", { status: 404 });
    } catch (error) {
      logger.error("Request failed", error as Error, {
        method: request.method,
        url: url.pathname,
      });

      return new Response(
        JSON.stringify({
          status: "error",
          message: (error as Error).message,
        }),
        { status: 500, headers: { "Content-Type": "application/json" } }
      );
    }
  },
};

function handleHealth(): Response {
  return new Response(
    JSON.stringify({
      status: "healthy",
      version: "1.0.0",
      timestamp: new Date().toISOString(),
      providers: {
        cloudflare: "available",
        groq: "available",
        huggingface: "available",
      },
    }),
    { headers: { "Content-Type": "application/json" } }
  );
}

async function handleTelegramWebhook(request: Request, env: Env, logger: ReturnType<typeof createLogger>): Promise<Response> {
  const secretToken = request.headers.get("X-Telegram-Bot-Api-Secret-Token");
  
  if (secretToken !== env.TELEGRAM_SECRET) {
    logger.warn("Unauthorized webhook attempt", { ip: request.headers.get("CF-Connecting-IP") });
    return new Response("Unauthorized", { status: 401 });
  }

  const update: TelegramUpdate = await request.json();

  if (!update.message?.text) {
    return new Response("OK", { status: 200 });
  }

  const chatId = update.message.chat.id;
  const text = update.message.text;

  logger.info("Received Telegram message", {
    chat_id: chatId,
    text_length: text.length,
  });

  if (text.startsWith("/start")) {
    return new Response(
      JSON.stringify({
        success: true,
        message: "Welcome! Send me a mathematical problem and I'll create an animated video explanation.",
      }),
      { headers: { "Content-Type": "application/json" } }
    );
  }

  if (text.startsWith("/help")) {
    return new Response(
      JSON.stringify({
        success: true,
        message: "Just describe any mathematical problem and I'll visualize it as an animated video!",
      }),
      { headers: { "Content-Type": "application/json" } }
    );
  }

  const jobId = crypto.randomUUID();
  logger.info("Creating job", { job_id: jobId, chat_id: chatId });

  return new Response(
    JSON.stringify({
      success: true,
      message: "Your video is being generated! This usually takes 1-5 minutes.",
      job_id: jobId,
    }),
    { headers: { "Content-Type": "application/json" } }
  );
}

async function handleAPI(request: Request, env: Env, url: URL, logger: ReturnType<typeof createLogger>): Promise<Response> {
  const method = request.method;

  if (method === "POST" && url.pathname === "/api/v1/jobs") {
    const body = await request.json();
    
    logger.info("Creating job via API", {
      session_id: body.session_id,
      problem_length: body.problem_text?.length,
    });

    const jobId = crypto.randomUUID();

    return new Response(
      JSON.stringify({
        job_id: jobId,
        status: "queued",
        message: "Job queued for processing",
        created_at: new Date().toISOString(),
      }),
      { headers: { "Content-Type": "application/json" } }
    );
  }

  if (method === "GET" && url.pathname.startsWith("/api/v1/jobs/")) {
    const jobId = url.pathname.split("/").pop();
    
    logger.debug("Getting job status", { job_id: jobId });

    return new Response(
      JSON.stringify({
        job_id: jobId,
        status: "queued",
        status_message: "Waiting to be processed",
        created_at: new Date().toISOString(),
      }),
      { headers: { "Content-Type": "application/json" } }
    );
  }

  return new Response("Not Found", { status: 404 });
}
