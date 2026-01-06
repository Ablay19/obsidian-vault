// Simplified Worker using only Cloudflare Workers AI
export default {
    async fetch(request, env) {
        const url = new URL(request.url);
        
        // Health check
        if (url.pathname === '/health') {
            return new Response(JSON.stringify({
                status: 'ok',
                timestamp: new Date().toISOString(),
                provider: 'cloudflare',
                model: '@cf/meta/llama-3.3-70b-instruct',
                environment: env.ENVIRONMENT || 'development',
                ai_available: !!env.AI
            }), {
                headers: { 'Content-Type': 'application/json' }
            });
        }
        
        // AI endpoint - using only Workers AI
        if (request.method === 'POST' && url.pathname === '/ai') {
            try {
                const body = await request.json();
                
                // Validate JSON payload
                if (!body || typeof body !== 'object') {
                    return new Response(JSON.stringify({
                        success: false,
                        error: 'Invalid JSON payload'
                    }), {
                        status: 400,
                        headers: { 'Content-Type': 'application/json' }
                    });
                }
                
                const { prompt } = body;
                
                if (!prompt || typeof prompt !== 'string' || prompt.trim().length === 0) {
                    return new Response(JSON.stringify({
                        success: false,
                        error: 'Valid prompt string is required'
                    }), {
                        status: 400,
                        headers: { 'Content-Type': 'application/json' }
                    });
                }
                
                // Sanitize prompt
                const sanitizedPrompt = prompt
                    .replace(/[\x00-\x1F\x7F]/g, '')
                    .substring(0, 4000);

                if (!env.AI) {
                    return new Response(JSON.stringify({
                        success: false,
                        error: 'Workers AI not available. Please configure AI binding.'
                    }), {
                        status: 503,
                        headers: { 'Content-Type': 'application/json' }
                    });
                }

                // Use Cloudflare Workers AI
                const response = await env.AI.run('@cf/meta/llama-3.3-70b-instruct', {
                    prompt: sanitizedPrompt,
                    max_tokens: 1000,
                    temperature: 0.7
                });
                
                return new Response(JSON.stringify({
                    success: true,
                    provider: 'cloudflare',
                    model: '@cf/meta/llama-3.3-70b-instruct',
                    response: response.response
                }), {
                    headers: { 
                        'Content-Type': 'application/json',
                        'x-ai-provider': 'cloudflare',
                        'x-model': '@cf/meta/llama-3.3-70b-instruct'
                    }
                });
            } catch (error) {
                return new Response(JSON.stringify({
                    success: false,
                    error: error.message
                }), {
                    status: 500,
                    headers: { 'Content-Type': 'application/json' }
                });
            }
        }
        
        // Telegram webhook (simplified)
        if (request.method === 'POST' && url.pathname === '/webhook') {
            try {
                const update = await request.json();
                console.log('Webhook received:', JSON.stringify(update));
                
                // Simple echo response for testing
                return new Response(JSON.stringify({
                    ok: true,
                    message: 'Webhook received',
                    timestamp: new Date().toISOString()
                }), {
                    status: 200,
                    headers: { 'Content-Type': 'application/json' }
                });
            } catch (error) {
                console.error('Webhook error:', error);
                return new Response(JSON.stringify({
                    ok: false,
                    error: error.message
                }), {
                    status: 500,
                    headers: { 'Content-Type': 'application/json' }
                });
            }
        }
        
        // Default response
        return new Response('Obsidian Bot Worker - Cloudflare AI Active', {
            status: 200,
            headers: { 'Content-Type': 'text/plain' }
        });
    }
};