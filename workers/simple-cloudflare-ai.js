// Simple Cloudflare Workers AI Proxy
export default {
  async fetch(request, env, ctx) {
    // Handle CORS
    if (request.method === 'OPTIONS') {
      return new Response(null, {
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type',
        }
      });
    }
    
    const url = new URL(request.url);
    
    // Health check
    if (url.pathname === '/health') {
      return new Response('OK', { status: 200 });
    }
    
    // AI proxy endpoint
    if (request.method === 'POST' && url.pathname === '/ai/proxy/cloudflare') {
      try {
        const prompt = await request.text();
        
        // Check if AI binding is available
        if (!env.AI) {
          return new Response(JSON.stringify({
            success: false,
            response: 'AI binding not configured. Please bind AI in wrangler.toml'
          }), {
            status: 500,
            headers: { 'Content-Type': 'application/json' }
          });
        }
        
        // Use Cloudflare Workers AI
        const response = await env.AI.run('@cf/meta/llama-3-8b-instruct', {
          prompt: prompt
        });
        
        return new Response(JSON.stringify({
          success: true,
          response: response.response
        }), {
          headers: {
            'Content-Type': 'application/json',
            'x-ai-provider': 'cloudflare',
            'x-model': '@cf/meta/llama-3-8b-instruct'
          }
        });
        
      } catch (error) {
        console.error('AI generation error:', error);
        return new Response(JSON.stringify({
          success: false,
          error: error.message
        }), {
          status: 500,
          headers: { 'Content-Type': 'application/json' }
        });
      }
    }
    
    // AI test endpoint
    if (url.pathname === '/ai-test') {
      const hasAI = !!env.AI;
      let models = null;
      
      if (hasAI) {
        try {
          models = await env.AI.list();
        } catch (e) {
          console.error('Failed to list models:', e);
        }
      }
      
      return new Response(JSON.stringify({
        hasAIBinding: hasAI,
        availableModels: models,
        error: hasAI ? null : 'AI binding not configured'
      }), {
        headers: { 'Content-Type': 'application/json' }
      });
    }
    
    return new Response('Not Found', { status: 404 });
  },
};