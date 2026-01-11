// Simplified Cloudflare Worker for Obsidian Bot
// Built-in Telegram handling with AI capabilities

import { BotUtils } from './bot-utils.js';

// Environment variable validation
function validateEnvironment(env) {
    const required = ['TELEGRAM_BOT_TOKEN'];
    const missing = required.filter(key => !env[key]);
    
    if (missing.length > 0) {
        throw new Error(`Missing required environment variables: ${missing.join(', ')}`);
}

// Configuration from wrangler.toml with validation
let BOT_TOKEN, WEBHOOK_SECRET;

// Simple in-memory state (production would use KV)
const userState = new Map();

// Clean old user state entries (older than 1 hour)
function cleanOldUserState() {
    const now = Date.now();
    const cutoff = now - 3600000; // 1 hour in milliseconds
    for (const [key, value] of userState.entries()) {
        if (value.timestamp < cutoff) {
            userState.delete(key);
        }
    }
}
    
    return {
        BOT_TOKEN: env.TELEGRAM_BOT_TOKEN,
        WEBHOOK_SECRET: env.WEBHOOK_SECRET || 'default-secret'
    };
}

// Configuration from wrangler.toml with validation
let BOT_TOKEN, WEBHOOK_SECRET;

// Simple in-memory state (production would use KV)
const userState = new Map();

// Utility functions

// Main worker handler
export default {
    async fetch(request, env) {
        // Validate environment on first request
        if (!BOT_TOKEN) {
            try {
                const config = validateEnvironment(env);
                BOT_TOKEN = config.BOT_TOKEN;
                WEBHOOK_SECRET = config.WEBHOOK_SECRET;
            } catch (error) {
                return new Response(JSON.stringify({
                    success: false,
                    error: error.message,
                    timestamp: new Date().toISOString()
                }), {
                    status: 500,
                    headers: { 'Content-Type': 'application/json' }
                });
            }
        }
        await BotUtils.log(`${request.method} ${request.url}`, 'info');

        try {
            // Health check endpoint
            if (request.url.includes('/health')) {
                return new Response(JSON.stringify({
                    status: 'ok',
                    timestamp: new Date().toISOString(),
                    environment: env.ENVIRONMENT || 'production',
                    ai_providers: {
                        gpt4: { available: env.GPT4_API_KEY ? true : false, model: 'gpt-4-turbo-preview' },
                        claude: { available: env.CLAUDE_API_KEY ? true : false, model: 'claude-3-sonnet-20240229' },
                        gemini: { available: env.GEMINI_API_KEYS ? true : false, model: 'gemini-pro' },
                        groq: { available: env.GROQ_API_KEY ? true : false, model: 'mixtral-8x7b-32768' },
                        cloudflare: { available: true, model: '@cf/meta/llama-2-7b-chat-int8' }
                    },
                    ai_proxy_features: ['caching', 'rate_limiting', 'cost_optimization', 'analytics', 'fallback_providers'],
                    worker_uptime: 'operational'
                }), {
                    headers: { 'Content-Type': 'application/json' }
                });
            }

            // AI API endpoint for external testing
            if (request.method === 'POST' && request.url.includes('/ai')) {
                const { prompt, provider } = await request.json();
                
                if (!prompt) {
                    return new Response(JSON.stringify({
                        success: false,
                        error: 'Prompt is required'
                    }), {
                        status: 400,
                        headers: { 'Content-Type': 'application/json' }
                    });
                }

                const result = await BotUtils.processAIRequest(prompt, env, provider);
                return new Response(JSON.stringify(result), {
                    status: result.success ? 200 : 500,
                    headers: { 'Content-Type': 'application/json' }
                });
            }

            // Metrics endpoint
            if (request.url.includes('/metrics')) {
                return new Response(JSON.stringify({
                    timestamp: new Date().toISOString(),
                    environment: env.ENVIRONMENT || 'production',
                    ai_providers: {
                        cloudflare: { available: !!env.AI }
                    },
                    requests_processed: userState.size,
                    worker_uptime: 'operational'
                }), {
                    headers: { 'Content-Type': 'application/json' }
                });
            }

            // Default response
            return new Response('Obsidian Bot Worker - Operational', {
                status: 200,
                headers: { 'Content-Type': 'text/plain' }
            });

        } catch (error) {
            await BotUtils.log(`Worker error: ${error.stack}`, 'error');
            
            return new Response(JSON.stringify({
                success: false,
                error: 'Internal server error',
                timestamp: new Date().toISOString()
            }), {
                status: 500,
                headers: { 'Content-Type': 'application/json' }
            });
        }
    }
};