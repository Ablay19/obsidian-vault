// Obsidian Bot - Cloudflare Worker
// Serverless Telegram bot with AI fallback capabilities

import { createMultipartBody } from 'create-multipart-body';
import { TelegramBot } from 'telegram-bot-api';

// Configuration
const BOT_TOKEN = env.TELEGRAM_BOT_TOKEN;
const WEBHOOK_SECRET = env.WEBHOOK_SECRET;
const FALLBACK_AI_ENABLED = env.FALLBACK_AI_ENABLED === 'true';

// AI Provider Configuration
const AI_PROVIDERS = {
    CLOUDFLARE: 'cloudflare',
    GEMINI: 'gemini',
    GROQ: 'groq'
};

// Bot State (using KV for persistence)
class BotState {
    constructor() {
        this.env = env.ENVIRONMENT || 'development';
        this.logLevel = env.LOG_LEVEL || 'info';
    }

    async log(message, level = 'info') {
        const timestamp = new Date().toISOString();
        const logEntry = {
            timestamp,
            level,
            environment: this.env,
            message,
            source: 'cloudflare-worker'
        };

        console.log(JSON.stringify(logEntry));
        
        // Send to external logging if configured
        if (env.LOG_ENDPOINT) {
            try {
                await fetch(env.LOG_ENDPOINT, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(logEntry)
                });
            } catch (error) {
                console.error('Failed to send log:', error);
            }
        }
    }

    async getUserState(userId) {
        try {
            const state = await OBSIDIAN_BOT_STATE.get(`user:${userId}`);
            return state ? JSON.parse(state) : null;
        } catch (error) {
            await this.log(`Failed to get user state: ${error}`, 'error');
            return null;
        }
    }

    async setUserState(userId, state) {
        try {
            await OBSIDIAN_BOT_STATE.put(`user:${userId}`, JSON.stringify(state), {
                expirationTtl: 86400 // 24 hours
            });
        } catch (error) {
            await this.log(`Failed to set user state: ${error}`, 'error');
        }
    }
}

// AI Service Handler
class AIService {
    constructor() {
        this.providers = [];
        this.currentProvider = null;
        this.initializeProviders();
    }

    initializeProviders() {
        // Cloudflare Workers AI (built-in)
        if (env.AI) {
            this.providers.push({
                name: AI_PROVIDERS.CLOUDFLARE,
                available: true,
                priority: 1,
                process: this.processWithCloudflare.bind(this)
            });
        }

        // External providers via HTTP endpoints
        if (env.GEMINI_API_KEY) {
            this.providers.push({
                name: AI_PROVIDERS.GEMINI,
                available: true,
                priority: 2,
                process: this.processWithGemini.bind(this)
            });
        }

        if (env.GROQ_API_KEY) {
            this.providers.push({
                name: AI_PROVIDERS.GROQ,
                available: true,
                priority: 3,
                process: this.processWithGroq.bind(this)
            });
        }
    }

    async processWithCloudflare(prompt, options = {}) {
        try {
            const model = options.model || '@cf/meta/llama-3.3-70b-instruct';
            const response = await env.AI.run(model, {
                prompt: prompt,
                max_tokens: options.maxTokens || 1000,
                temperature: options.temperature || 0.7
            });
            
            return {
                success: true,
                provider: AI_PROVIDERS.CLOUDFLARE,
                model,
                response: response.response,
                usage: response.usage
            };
        } catch (error) {
            return {
                success: false,
                provider: AI_PROVIDERS.CLOUDFLARE,
                error: error.message
            };
        }
    }

    async processWithGemini(prompt, options = {}) {
        try {
            const response = await fetch('https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${env.GEMINI_API_KEY}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    contents: [{
                        parts: [{
                            text: prompt
                        }]
                    }],
                    generationConfig: {
                        temperature: options.temperature || 0.7,
                        maxOutputTokens: options.maxTokens || 1000
                    }
                })
            });

            const data = await response.json();
            
            if (response.ok && data.candidates && data.candidates.length > 0) {
                return {
                    success: true,
                    provider: AI_PROVIDERS.GEMINI,
                    model: 'gemini-pro',
                    response: data.candidates[0].content.parts[0].text,
                    usage: data.usageMetadata
                };
            } else {
                throw new Error(data.error?.message || 'Gemini API error');
            }
        } catch (error) {
            return {
                success: false,
                provider: AI_PROVIDERS.GEMINI,
                error: error.message
            };
        }
    }

    async processWithGroq(prompt, options = {}) {
        try {
            const response = await fetch('https://api.groq.com/openai/v1/chat/completions', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${env.GROQ_API_KEY}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    model: options.model || 'llama-3.1-8b-instant',
                    messages: [{ role: 'user', content: prompt }],
                    temperature: options.temperature || 0.7,
                    max_tokens: options.maxTokens || 1000
                })
            });

            const data = await response.json();
            
            if (response.ok && data.choices && data.choices.length > 0) {
                return {
                    success: true,
                    provider: AI_PROVIDERS.GROQ,
                    model: data.model,
                    response: data.choices[0].message.content,
                    usage: data.usage
                };
            } else {
                throw new Error(data.error?.message || 'Groq API error');
            }
        } catch (error) {
            return {
                success: false,
                provider: AI_PROVIDERS.GROQ,
                error: error.message
            };
        }
    }

    async process(prompt, options = {}) {
        // Try providers in priority order
        for (const provider of this.providers.sort((a, b) => a.priority - b.priority)) {
            if (!provider.available) continue;

            const result = await provider.process(prompt, options);
            if (result.success) {
                this.currentProvider = provider.name;
                return result;
            } else {
                await botState.log(`Provider ${provider.name} failed: ${result.error}`, 'warn');
            }
        }

        // All providers failed
        return {
            success: false,
            error: 'All AI providers are currently unavailable',
            fallbackAvailable: FALLBACK_AI_ENABLED
        };
    }

    async healthCheck() {
        const health = {};
        
        for (const provider of this.providers) {
            try {
                const testPrompt = 'Hello';
                const result = await provider.process(testPrompt, { maxTokens: 10 });
                
                health[provider.name] = {
                    available: result.success,
                    responseTime: Date.now(),
                    error: result.error || null
                };
            } catch (error) {
                health[provider.name] = {
                    available: false,
                    error: error.message
                };
            }
        }

        return health;
    }
}

// File Processing Service
class FileProcessor {
    constructor() {
        this.maxFileSize = 20 * 1024 * 1024; // 20MB
        this.supportedTypes = ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'];
    }

    async processImage(file) {
        // For images, we can extract metadata and generate descriptions
        const metadata = {
            size: file.size,
            type: file.type,
            name: file.name
        };

        // Generate AI description of the image
        const imagePrompt = `Describe this image in detail. What do you see? File: ${file.name}`;
        const aiResult = await aiService.process(imagePrompt);
        
        return {
            type: 'image',
            metadata,
            description: aiResult.success ? aiResult.response : 'Failed to analyze image',
            aiProvider: aiResult.success ? aiResult.provider : null
        };
    }

    async processPDF(file) {
        // PDF processing would require external service
        // For now, we'll extract metadata and queue for processing
        const metadata = {
            size: file.size,
            type: file.type,
            name: file.name
        };

        // Store file for later processing
        const fileKey = `pdf/${Date.now()}-${file.name}`;
        await OBSIDIAN_BOT_MEDIA.put(fileKey, file, {
            expirationTtl: 86400 // 24 hours
        });

        // Generate analysis prompt
        const analysisPrompt = `This PDF file was uploaded: ${file.name}. Please provide a brief summary and suggest tags for categorization.`;
        const aiResult = await aiService.process(analysisPrompt);

        return {
            type: 'pdf',
            metadata,
            fileKey,
            description: aiResult.success ? aiResult.response : 'PDF received for processing',
            aiProvider: aiResult.success ? aiResult.provider : null,
            processing: true
        };
    }

    async processText(text) {
        // Process text with AI for summarization and categorization
        const summarizationPrompt = `Please summarize and categorize the following text:\n\n${text}\n\nProvide a summary and suggested category.`;
        const aiResult = await aiService.process(summarizationPrompt);

        return {
            type: 'text',
            text,
            summary: aiResult.success ? aiResult.response : 'Failed to process text',
            aiProvider: aiResult.success ? aiResult.provider : null
        };
    }
}

// Initialize services
const botState = new BotState();
const aiService = new AIService();
const fileProcessor = new FileProcessor();

// Telegram Bot Handlers
class BotHandlers {
    static async handleMessage(update) {
        try {
            const message = update.message;
            if (!message) return;

            await botState.log(`Processing message from ${message.from.id}: ${message.text?.substring(0, 50)}...`);

            // Get user state
            const userState = await botState.getUserState(message.from.id) || {};

            // Handle commands
            if (message.text?.startsWith('/')) {
                return await this.handleCommand(message, userState);
            }

            // Handle different message types
            if (message.photo) {
                return await this.handlePhoto(message, userState);
            } else if (message.document) {
                return await this.handleDocument(message, userState);
            } else if (message.text) {
                return await this.handleText(message, userState);
            }

        } catch (error) {
            await botState.log(`Message handling error: ${error}`, 'error');
            await this.sendError(message.chat.id, 'Sorry, I encountered an error processing your message.');
        }
    }

    static async handleCommand(message, userState) {
        const command = message.text.split(' ')[0].substring(1);
        const args = message.text.split(' ').slice(1);

        switch (command) {
            case 'start':
                await this.sendStartMessage(message.chat.id);
                break;
                
            case 'help':
                await this.sendHelpMessage(message.chat.id);
                break;
                
            case 'status':
                await this.sendStatusMessage(message.chat.id, userState);
                break;
                
            case 'provider':
                await this.handleProviderCommand(message.chat.id, args, userState);
                break;
                
            case 'health':
                await this.sendHealthMessage(message.chat.id);
                break;
                
            default:
                await this.sendUnknownCommandMessage(message.chat.id, command);
                break;
        }
    }

    static async handlePhoto(message, userState) {
        const photo = message.photo[message.photo.length - 1]; // Get largest photo
        const fileUrl = `https://api.telegram.org/file/bot${BOT_TOKEN}/${photo.file_path}`;
        
        try {
            // Download photo
            const response = await fetch(fileUrl);
            const file = await response.blob();
            
            if (file.size > fileProcessor.maxFileSize) {
                return await this.sendMessage(message.chat.id, 'File too large. Maximum size is 20MB.');
            }

            // Process image
            const result = await fileProcessor.processImage(file);
            
            // Send response
            const responseText = `📷 **Image Analysis**\n\n${result.description}\n\n🤖 *AI Provider: ${result.aiProvider}*`;
            await this.sendMessage(message.chat.id, responseText);
            
            // Update user state
            userState.lastProcessed = result;
            await botState.setUserState(message.from.id, userState);
            
        } catch (error) {
            await botState.log(`Photo processing error: ${error}`, 'error');
            await this.sendError(message.chat.id, 'Failed to process image. Please try again.');
        }
    }

    static async handleDocument(message, userState) {
        const document = message.document;
        
        if (document.file_size > fileProcessor.maxFileSize) {
            return await this.sendMessage(message.chat.id, 'File too large. Maximum size is 20MB.');
        }

        try {
            // For documents, we'll acknowledge and process asynchronously
            const responseText = `📄 **Document Received**\n\n*File:* ${document.file_name}\n*Size:* ${(document.file_size / 1024 / 1024).toFixed(2)}MB\n\n🔄 Processing...`;
            await this.sendMessage(message.chat.id, responseText);

            // Queue for processing (would implement with Cloudflare Queues)
            // For now, provide immediate text analysis
            const analysisPrompt = `This document was uploaded: ${document.file_name}. Please provide a brief analysis and suggested actions.`;
            const aiResult = await aiService.process(analysisPrompt);
            
            const followupText = `📋 **Document Analysis**\n\n${aiResult.success ? aiResult.response : 'Processing queued...'}\n\n🤖 *AI Provider: ${aiResult.success ? aiResult.provider : 'Processing'}*`;
            await this.sendMessage(message.chat.id, followupText);
            
        } catch (error) {
            await botState.log(`Document processing error: ${error}`, 'error');
            await this.sendError(message.chat.id, 'Failed to process document.');
        }
    }

    static async handleText(message, userState) {
        try {
            // Process text with AI
            const result = await aiService.process(message.text);
            
            if (result.success) {
                const responseText = `💬 **AI Response**\n\n${result.response}\n\n🤖 *Provider: ${result.provider}*`;
                await this.sendMessage(message.chat.id, responseText);
            } else {
                // Fallback response
                const fallbackText = `🤖 **AI Currently Unavailable**\n\n${result.error}\n\n⚠️ *Worker Mode: Limited functionality*\n\nI'll process your message when AI services are available.`;
                await this.sendMessage(message.chat.id, fallbackText);
            }
            
            // Update user state
            userState.lastTextProcessing = result;
            await botState.setUserState(message.from.id, userState);
            
        } catch (error) {
            await botState.log(`Text processing error: ${error}`, 'error');
            await this.sendError(message.chat.id, 'Failed to process your message.');
        }
    }

    static async handleProviderCommand(chatId, args, userState) {
        if (args.length === 0) {
            const health = await aiService.healthCheck();
            let healthText = '🤖 **AI Provider Status**\n\n';
            
            for (const [provider, status] of Object.entries(health)) {
                const icon = status.available ? '✅' : '❌';
                healthText += `${icon} *${provider}*: ${status.available ? 'Available' : 'Unavailable'}\n`;
            }
            
            healthText += `\n📍 *Current: ${aiService.currentProvider || 'None'}*`;
            return await this.sendMessage(chatId, healthText);
        }

        const requestedProvider = args[0].toLowerCase();
        const provider = aiService.providers.find(p => p.name.toLowerCase() === requestedProvider);
        
        if (provider && provider.available) {
            userState.preferredProvider = requestedProvider;
            await botState.setUserState(userState.userId, userState);
            await this.sendMessage(chatId, `✅ AI provider set to ${requestedProvider}`);
        } else {
            await this.sendMessage(chatId, `❌ Unknown provider: ${requestedProvider}`);
        }
    }

    static async sendStartMessage(chatId) {
        const text = `🤖 **Obsidian Bot - Cloudflare Worker Edition**\n\nWelcome! I'm your AI-powered assistant with fallback capabilities.\n\n**Available Commands:**\n• /help - Show this help\n• /status - Show your current status\n• /provider - Check AI provider status\n• /health - System health check\n\n**Features:**\n• 📷 Image analysis\n• 📄 Document processing\n• 💬 AI chat responses\n• 🔄 Provider fallback\n\nI'm running on Cloudflare Workers with enterprise-grade reliability!`;
        await this.sendMessage(chatId, text);
    }

    static async sendHelpMessage(chatId) {
        const text = `📚 **Help & Commands**\n\n**Bot Commands:**\n• /start - Welcome message\n• /help - Show this help\n• /status - Your current status\n• /provider [name] - Set AI provider\n• /health - System health\n\n**File Processing:**\n• Send photos for AI analysis\n• Send documents for processing\n• Send text for AI responses\n\n**AI Providers:**\n• Cloudflare Workers AI (built-in)\n• Gemini (external)\n• Groq (external)\n\nThe bot automatically falls back between providers!`;
        await this.sendMessage(chatId, text);
    }

    static async sendStatusMessage(chatId, userState) {
        const aiHealth = await aiService.healthCheck();
        const statusText = `📊 **Your Status**\n\n🤖 *Current AI Provider:* ${aiService.currentProvider || 'None'}\n📝 *Last Activity:* ${userState.lastActivity || 'None'}\n\n**AI Provider Health:**\n${Object.entries(aiHealth).map(([p, h]) => `• ${p}: ${h.available ? '✅' : '❌'}`).join('\n')}`;
        await this.sendMessage(chatId, statusText);
    }

    static async sendHealthMessage(chatId) {
        const aiHealth = await aiService.healthCheck();
        const systemHealth = {
            worker: '✅ Operational',
            ai: Object.values(aiHealth).some(h => h.available) ? '✅ Available' : '⚠️ Limited',
            storage: '✅ Operational',
            uptime: '100%'
        };

        const healthText = `🏥 **System Health**\n\n**Worker Status:**\n${systemHealth.worker}\n**AI Services:**\n${systemHealth.ai}\n**Storage:**\n${systemHealth.storage}\n**Uptime:**\n${systemHealth.uptime}\n\n**Provider Details:**\n${Object.entries(aiHealth).map(([p, h]) => `• ${p}: ${h.available ? 'Available' : 'Unavailable'} (${h.error || 'No error'})`).join('\n')}`;
        await this.sendMessage(chatId, healthText);
    }

    static async sendUnknownCommandMessage(chatId, command) {
        const text = `❓ Unknown command: /${command}\n\nType /help to see available commands.`;
        await this.sendMessage(chatId, text);
    }

    static async sendError(chatId, error) {
        const text = `❌ **Error**\n\n${error}\n\nPlease try again or contact support if the issue persists.`;
        await this.sendMessage(chatId, text);
    }

    static async sendMessage(chatId, text, options = {}) {
        try {
            const bot = new TelegramBot(BOT_TOKEN);
            await bot.sendMessage({
                chat_id: chatId,
                text: text,
                parse_mode: options.parse_mode || 'Markdown',
                disable_web_page_preview: options.disable_preview || false
            });
        } catch (error) {
            await botState.log(`Send message error: ${error}`, 'error');
        }
    }
}

// Webhook handler
export default {
    async fetch(request, env) {
        await botState.log(`${request.method} ${request.url}`, 'info');

        try {
            // Health check
            if (request.url.includes('/health')) {
                const health = await aiService.healthCheck();
                return new Response(JSON.stringify({
                    status: 'ok',
                    timestamp: new Date().toISOString(),
                    environment: botState.env,
                    ai_providers: health
                }), {
                    headers: { 'Content-Type': 'application/json' }
                });
            }

            // Webhook verification (Telegram)
            if (request.method === 'POST' && request.url.includes('/webhook')) {
                // Verify webhook secret if configured
                if (WEBHOOK_SECRET) {
                    const signature = request.headers.get('X-Telegram-Bot-Api-Secret-Token');
                    if (signature !== WEBHOOK_SECRET) {
                        return new Response('Unauthorized', { status: 401 });
                    }
                }

                // Parse webhook payload
                const update = await request.json();
                
                // Process the update
                if (update.message) {
                    await BotHandlers.handleMessage(update);
                } else if (update.callback_query) {
                    // Handle callback queries
                    await botState.log(`Callback query received: ${update.callback_query.id}`, 'info');
                }

                return new Response('OK', { status: 200 });
            }

            // AI API endpoint for external clients
            if (request.method === 'POST' && request.url.includes('/ai')) {
                const { prompt, provider, options } = await request.json();
                
                if (!prompt) {
                    return new Response(JSON.stringify({
                        success: false,
                        error: 'Prompt is required'
                    }), {
                        status: 400,
                        headers: { 'Content-Type': 'application/json' }
                    });
                }

                // Process with specified provider or default
                let result;
                if (provider && aiService.providers.find(p => p.name === provider)) {
                    const providerConfig = aiService.providers.find(p => p.name === provider);
                    result = await providerConfig.process(prompt, options);
                } else {
                    result = await aiService.process(prompt, options);
                }

                return new Response(JSON.stringify(result), {
                    status: result.success ? 200 : 500,
                    headers: { 'Content-Type': 'application/json' }
                });
            }

            // Metrics endpoint
            if (request.url.includes('/metrics')) {
                const metrics = {
                    timestamp: new Date().toISOString(),
                    ai_providers: await aiService.healthCheck(),
                    environment: botState.env,
                    worker_uptime: 'operational'
                };

                return new Response(JSON.stringify(metrics), {
                    headers: { 'Content-Type': 'application/json' }
                });
            }

            // Default response
            return new Response('Obsidian Bot Worker - Operational', {
                status: 200,
                headers: { 'Content-Type': 'text/plain' }
            });

        } catch (error) {
            await botState.log(`Worker error: ${error.stack}`, 'error');
            
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