# AI Platform Monitoring Setup

This directory contains the monitoring configuration for the AI Platform using Prometheus and Grafana.

## Components

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboards
- **Node Exporter**: Host system metrics
- **cAdvisor**: Container metrics

## Quick Start

1. Start the monitoring stack:
```bash
cd monitoring
docker-compose up -d
```

2. Access Grafana at http://localhost:3000 (admin/admin)

3. Import the dashboard from `grafana/dashboards/ai-platform-dashboard.json`

## Services

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
- **Node Exporter**: http://localhost:9100
- **cAdvisor**: http://localhost:8081

## Metrics Collection

The system collects metrics from:
- API Gateway (Go application)
- AI Worker (Cloudflare Worker)
- Manim Renderer (Python application)
- Host system metrics
- Container metrics

## Dashboard

The included Grafana dashboard shows:
- Request rates and response times
- Error rates
- Health status
- Resource usage (CPU, memory)
- Container metrics

## Configuration

- `prometheus/prometheus.yml`: Prometheus scrape configuration
- `grafana/dashboards/`: Grafana dashboard definitions
- `docker-compose.yml`: Monitoring stack orchestration

## Adding Custom Metrics

To add custom metrics to your services:

### Go Applications
Use the `prometheus/client_golang` library:
```go
import "github.com/prometheus/client_golang/prometheus"

var requestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)
```

### JavaScript Workers
Use Cloudflare's metrics API or expose metrics via HTTP endpoints.

## Alerting

Configure alerts in Prometheus for:
- Service downtime
- High error rates
- Resource exhaustion
- Performance degradation