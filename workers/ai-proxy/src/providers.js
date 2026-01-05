// AI Provider Management for Cloudflare Workers
export class AIProviders {
  constructor(env) {
    this.env = env;
    this.providers = {
      gemini: {
        name: 'gemini',
        endpoint: 'https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent',
        apiKey: env.GEMINI_API_KEY,
        model: 'gemini-pro',
        maxTokens: 8192,
        costPerToken: 0.000125,
        latency: 800,
        enabled: env.GEMINI_API_KEY ? true : false
      },
      groq: {
        name: 'groq',
        endpoint: 'https://api.groq.com/openai/v1/chat/completions',
        apiKey: env.GROQ_API_KEY,
        model: 'mixtral-8x7b-32768',
        maxTokens: 4096,
        costPerToken: 0.00005,
        latency: 400,
        enabled: env.GROQ_API_KEY ? true : false
      },
      claude: {
        name: 'claude',
        endpoint: 'https://api.anthropic.com/v1/messages',
        apiKey: env.CLAUDE_API_KEY,
        model: 'claude-3-sonnet-20240229',
        maxTokens: 100000,
        costPerToken: 0.0008,
        latency: 1200,
        enabled: env.CLAUDE_API_KEY ? true : false
      },
      gpt4: {
        name: 'gpt4',
        endpoint: 'https://api.openai.com/v1/chat/completions',
        apiKey: env.OPENAI_API_KEY,
        model: 'gpt-4-turbo-preview',
        maxTokens: 8192,
        costPerToken: 0.00003,
        latency: 600,
        enabled: env.OPENAI_API_KEY ? true : false
      }
    };
  }
  
  getProvider(name) {
    return this.providers[name] || null;
  }
  
  getAllProviders() {
    return Object.values(this.providers);
  }
  
  getEnabledProviders() {
    return Object.values(this.providers).filter(p => p.enabled);
  }
}

export default AIProviders;