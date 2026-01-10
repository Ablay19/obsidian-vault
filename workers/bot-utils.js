// Shared BotUtils class for Obsidian Bot Workers
// Eliminates code duplication across worker implementations

export class BotUtils {
    static sanitizeInput(text) {
        if (typeof text !== 'string') return '';
        // Remove potentially harmful characters
        return text
            .replace(/[\x00-\x1F\x7F]/g, '') // Remove control characters
            .replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '') // Remove more control chars
            .substring(0, 4000); // Limit length
    }

    static async log(message, level = 'info') {
        console.log(JSON.stringify({
            timestamp: new Date().toISOString(),
            level,
            message,
            source: 'obsidian-bot-worker'
        }));
    }

    static async sendTelegramMessage(botToken, chatId, text) {
        try {
            // Sanitize inputs
            const sanitizedText = this.sanitizeInput(text);
            const sanitizedChatId = this.sanitizeInput(chatId.toString());
            
            const response = await fetch(`https://api.telegram.org/bot${botToken}/sendMessage`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    chat_id: sanitizedChatId,
                    text: sanitizedText,
                    parse_mode: 'Markdown'
                })
            });

            if (!response.ok) {
                throw new Error(`Failed to send message: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            await this.log(`Send message error: ${error}`, 'error');
            throw error;
        }
    }

    static async processAIRequest(prompt, env, provider = 'cloudflare') {
        try {
            // Use Cloudflare Workers AI for all requests
            if (env.AI) {
                const response = await env.AI.run('@cf/meta/llama-2-7b-chat-int8', {
                    prompt: this.sanitizeInput(prompt),
                    max_tokens: 1000,
                    temperature: 0.7
                });

                return {
                    success: true,
                    provider: 'cloudflare',
                    model: '@cf/meta/llama-2-7b-chat-int8',
                    response: response.response
                };
            }

            // Fallback if no AI binding
            return {
                success: false,
                error: 'Cloudflare AI not configured',
                provider: provider
            };

        } catch (error) {
            return {
                success: false,
                provider: provider,
                error: error.message
            };
        }
    }

    static formatAIResponse(result) {
        if (!result.success) {
            return `AI Currently Unavailable\\n\\nError: ${result.error}\\n\\nWorker Mode: Basic functionality\\n\\nPlease try again or contact support if the issue persists.`;
        }

        return `AI Response\\n\\n${result.response}\\n\\nProvider: ${result.provider}\\nModel: ${result.model}`;
    }
}


