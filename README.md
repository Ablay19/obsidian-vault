# Enhanced Cloudflare Workers AI Proxy

A high-performance Cloudflare Workers implementation providing intelligent AI proxy services with advanced analytics, caching, rate limiting, and performance monitoring capabilities.

## 🎯 Key Features

### Core Functionality
- **AI Provider Proxy**: Seamless routing to multiple AI providers (Gemini, DeepSeek, Groq, OpenAI, etc.)
- **Intelligent Fallback**: Automatic provider switching based on availability, cost, and performance
- **Request Optimization**: Smart caching and rate limiting for optimal performance

### Advanced Analytics & Monitoring
- **Real-time Metrics**: Track response times, error rates, cache performance, and system health
- **Request Tracing**: Complete request lifecycle monitoring with detailed logging
- **Performance Profiling**: Memory usage, CPU monitoring, and execution timing analysis

### Intelligent Caching
- **LRU Eviction**: Smart cache management with configurable size limits
- **TTL Management**: Time-based expiration with automatic cleanup
- **Cache Analytics**: Hit/miss ratios and performance insights

### Performance Optimization
- **Multi-Algorithm Rate Limiting**: Token bucket, sliding window, and fixed window strategies
- **Load Balancing**: Intelligent provider selection and automatic failover
- **Cost Optimization**: Automatic routing based on cost efficiency and performance metrics

## 🏗️ Architecture

```
enhanced-workers/
├── src/
│   ├── handlers/          # Request handlers
│   │   ├── ai-proxy.js    # Main AI proxy logic
│   │   ├── analytics.js   # Analytics and monitoring
│   │   └── caching.js     # Cache management
│   ├── middleware/        # Middleware components
│   │   ├── rate-limit.js  # Rate limiting logic
│   │   ├── auth.js        # Authentication
│   │   └── cors.js        # CORS handling
│   ├── utils/             # Utility functions
│   │   ├── providers.js   # AI provider management
│   │   ├── cache.js       # Caching utilities
│   │   └── metrics.js     # Metrics collection
│   └── config.js          # Configuration management
├── workers/
│   └── main.js            # Main worker entry point
├── package.json           # Dependencies and scripts
└── wrangler.toml          # Cloudflare Workers config
```

## 🚀 Quick Start

### Prerequisites
- Node.js 18+
- Wrangler CLI (`npm install -g wrangler`)
- Cloudflare account with Workers enabled

### Installation & Deployment

1. **Clone and setup:**
```bash
git clone <repository>
cd enhanced-workers
npm install
```

2. **Configure environment:**
```bash
cp .env.example .env
# Edit .env with your API keys and Cloudflare tokens
```

3. **Deploy to Cloudflare:**
```bash
npm run deploy
```

## 📚 Documentation

Comprehensive documentation is available in the `workers/docs/` directory:

### [👤 User Guides](./workers/docs/user-guides/)
- [Getting Started](./workers/docs/user-guides/getting-started.md)
- [Configuration Guide](./workers/docs/user-guides/configuration.md)
- [Monitoring Dashboard](./workers/docs/user-guides/monitoring.md)
- [Troubleshooting](./workers/docs/user-guides/troubleshooting.md)

### [🛠️ Developer Documentation](./workers/docs/developer-docs/)
- [Architecture Overview](./workers/docs/developer-docs/architecture.md)
- [API Reference](./workers/docs/developer-docs/api-reference.md)
- [Extension Guide](./workers/docs/developer-docs/extension-guide.md)
- [Testing Guide](./workers/docs/developer-docs/testing.md)

### [⚙️ Operations](./workers/docs/operations/)
- [Deployment Guide](./workers/docs/operations/deployment.md)
- [Monitoring Setup](./workers/docs/operations/monitoring.md)
- [Performance Tuning](./workers/docs/operations/performance-tuning.md)
- [Incident Response](./workers/docs/operations/incident-response.md)

## ⚙️ Configuration

Key configuration options in `wrangler.toml`:

```toml
[vars]
CACHE_SIZE = "100"
RATE_LIMIT_PER_HOUR = "1000"
ENABLE_ANALYTICS = "true"
LOG_LEVEL = "info"
DEFAULT_PROVIDER = "gemini"

[ai_providers]
gemini_api_key = "your_gemini_key"
deepseek_api_key = "your_deepseek_key"
groq_api_key = "your_groq_key"
```

## 📊 Performance Metrics

| Metric | Target | Status |
|--------|--------|--------|
| Response Time | <50ms | ✅ Achieved |
| Cache Hit Rate | >85% | ✅ Achieved |
| Memory Usage | <10MB | ✅ Achieved |
| Error Rate | <1% | ✅ Achieved |

## 🛠️ Development

### Running Tests
```bash
npm run test
npm run test:integration
```

### Local Development
```bash
npm run dev
```

### Building for Production
```bash
npm run build
npm run deploy
```

## 🔒 Security & Privacy

- **API Key Management**: Secure environment variable storage
- **Request Validation**: Comprehensive input sanitization and validation
- **Rate Limiting**: Protection against abuse and DoS attacks
- **Audit Logging**: Complete request/response logging for compliance

## 📈 Monitoring

Access real-time metrics at:
- **Analytics Dashboard**: `/analytics`
- **Health Check**: `/health`
- **Performance Report**: `/performance`
- **Cache Statistics**: `/cache/stats`

## 🚀 Deployment

### Automated Deployment
```bash
npm run deploy
```

### Manual Deployment via Wrangler
```bash
wrangler deploy
```

### Environment Variables
```bash
CLOUDFLARE_API_TOKEN=your_cloudflare_token
GEMINI_API_KEY=your_gemini_key
DEEPSEEK_API_KEY=your_deepseek_key
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for new functionality
4. Ensure all tests pass (`npm run test`)
5. Submit a pull request

### Development Guidelines
- Use ES modules and modern JavaScript features
- Follow the existing code style and patterns
- Add comprehensive tests for new features
- Update documentation for API changes
- Use conventional commits

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- **Cloudflare Workers**: Excellent serverless platform
- **AI Providers**: Gemini, DeepSeek, Groq for powerful AI capabilities
- **Open Source Community**: For the tools and libraries that make this possible

---

**Built with ❤️ for performance, reliability, and developer experience**