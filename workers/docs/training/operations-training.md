# Operations Training

## 🔧 Operations Training Overview

This training program provides comprehensive onboarding for operations teams working with enhanced Cloudflare Workers. It covers deployment procedures, monitoring setup, incident response, and performance optimization.

## 📚 Training Modules

### Module 1: Platform Overview

#### Learning Objectives
- Understand the enhanced workers platform
- Identify key operational components
- Learn about deployment environments
- Understand monitoring and alerting

#### Session 1: Platform Architecture

```yaml
duration: 2 hours
format: Lecture + Interactive Demo

topics:
  - Enhanced Workers Platform Overview
  - Operational Components:
    - Cloudflare Workers
    - Cloudflare Analytics
    - Durable Objects (if applicable)
    - KV Storage
    - R2 Storage (if applicable)
  - Deployment Environments:
    - Development
    - Staging
    - Production
  - Monitoring Stack:
    - Metrics Collection
    - Log Aggregation
    - Alerting System
    - Dashboard Setup

activities:
  - Platform walkthrough
  - Environment configuration review
  - Monitoring dashboard tour
  - Q&A session
```

#### Session 2: Deployment Pipeline

```yaml
duration: 2 hours
format: Hands-on Lab

topics:
  - CI/CD Pipeline Overview
  - Wrangler CLI Usage
  - Environment Configuration
  - Deployment Strategies:
    - Blue-Green Deployment
    - Canary Deployment
    - Rolling Deployment
  - Rollback Procedures

hands-on:
  1. Set up Wrangler CLI
     ```bash
     npm install -g wrangler
     wrangler login
     wrangler whoami
     ```

  2. Configure environments
     ```bash
     # Set up development environment
     wrangler deploy --env development

     # Set up staging environment
     wrangler deploy --env staging

     # Set up production environment
     wrangler deploy --env production
     ```

  3. Configure secrets
     ```bash
     wrangler secret put OPENAI_API_KEY --env production
     wrangler secret put ANTHROPIC_API_KEY --env production
     ```

  4. Deploy to production
     ```bash
     wrangler deploy --env production
     ```

exercises:
  - Deploy a worker to staging environment
  - Configure environment variables
  - Test deployment process
  - Perform a rollback
```

### Module 2: Monitoring and Alerting

#### Learning Objectives
- Set up monitoring dashboards
- Configure alerting rules
- Understand metrics and logs
- Master alert response procedures

#### Session 1: Monitoring Setup

```yaml
duration: 2 hours
format: Workshop

topics:
  - Cloudflare Analytics Setup
  - Custom Metrics Configuration
  - Dashboard Creation
  - Log Collection and Analysis

hands-on:
  - Set up Cloudflare Analytics
  ```bash
  # Enable analytics
  wrangler analytics enable

  # View analytics
  wrangler analytics dashboard

  # Export data
  wrangler analytics export --start 2024-01-01 --end 2024-01-31
  ```

  - Create custom dashboard
  ```javascript
  // Dashboard configuration
  const dashboardConfig = {
    title: 'Workers Operations Dashboard',
    widgets: [
      {
        type: 'line',
        title: 'Request Rate',
        query: 'sum(rate(requests_total[5m]))'
      },
      {
        type: 'line',
        title: 'Response Time',
        query: 'histogram_quantile(0.95, request_duration_seconds)'
      },
      {
        type: 'gauge',
        title: 'Error Rate',
        query: 'rate(errors_total[5m]) / rate(requests_total[5m])'
      }
    ]
  };
  ```

exercises:
  - Create a custom monitoring dashboard
  - Set up log aggregation
  - Configure metrics collection
  - Test monitoring data flow
```

#### Session 2: Alerting Configuration

```yaml
duration: 2 hours
format: Workshop

topics:
  - Alert Types and Severity Levels
  - Alert Threshold Configuration
  - Notification Channels:
    - Slack
    - PagerDuty
    - Email
    - SMS
  - Alert Escalation Rules

hands-on:
  - Configure Slack alerts
  ```yaml
  # slack-alert-config.yml
  alerts:
    - name: "High Error Rate"
      condition: "error_rate > 0.05"
      severity: "critical"
      channels:
        - slack
      slack:
        webhook: "https://hooks.slack.com/services/xxx"
        channel: "#alerts"
  ```

  - Configure PagerDuty integration
  ```javascript
  // PagerDuty alert configuration
  const pagerDutyConfig = {
    routingKey: process.env.PAGERDUTY_ROUTING_KEY,
    eventAction: 'trigger',
    payload: {
      summary: 'High error rate detected',
      severity: 'critical',
      source: 'cloudflare-workers',
      custom_details: {
        error_rate: '0.08',
        threshold: '0.05'
      }
    }
  };
  ```

exercises:
  - Create alert rules for different metrics
  - Configure multiple notification channels
  - Test alert delivery
  - Set up alert escalation
```

### Module 3: Incident Response

#### Learning Objectives
- Understand incident classification
- Master incident response procedures
- Learn communication protocols
- Conduct post-incident analysis

#### Session 1: Incident Detection and Triage

```yaml
duration: 2 hours
format: Simulation Exercise

topics:
  - Incident Severity Levels (P0-P3)
  - Detection Mechanisms
  - Triage Procedures
  - Escalation Matrix

simulation:
  scenario: "High error rate spike"

  steps:
    1. Detect incident
       - Monitor alerts firing
       - Verify incident severity
       - Classify incident type

    2. Initialize triage
       - Assign incident ID
       - Notify responders
       - Begin investigation

    3. Gather information
       - Check metrics
       - Review logs
       - Analyze traces

exercises:
  - Classify various incident scenarios
  - Practice triage procedures
  - Use incident response tools
  - Conduct initial investigation
```

#### Session 2: Incident Response Procedures

```yaml
duration: 3 hours
format: Role-playing Exercise

topics:
  - Response Team Roles and Responsibilities
  - Communication Protocols
  - Mitigation Strategies
  - Recovery Procedures

role_play:
  scenario: "Service outage"

  roles:
    - Incident Commander
    - Communications Lead
    - Technical Lead
    - Support Lead

  procedure:
    1. Declare incident
    2. Assemble response team
    3. Investigate root cause
    4. Implement mitigation
    5. Communicate updates
    6. Resolve incident
    7. Document postmortem

exercises:
  - Practice incident communication
  - Execute mitigation procedures
  - Coordinate with stakeholders
  - Manage incident timeline
```

#### Session 3: Post-Incident Analysis

```yaml
duration: 2 hours
format: Workshop

topics:
  - Postmortem Creation
  - Root Cause Analysis
  - Action Item Tracking
  - Continuous Improvement

hands-on:
  - Create postmortem
  ```yaml
  incident_id: INC-2024-001
  title: P0 Service Outage
  date: 2024-01-15
  severity: P0

  summary: |
    On January 15, 2024, at 14:30 UTC, we experienced a complete
    service outage lasting 45 minutes, affecting 100% of users.

  impact:
    users_affected: 100%
    duration: 45 minutes
    revenue_impact: $5000

  root_cause: |
    A recent deployment contained a configuration error that caused
    the provider manager to fail initialization, resulting in all
    requests failing.

  timeline:
    - "14:30 UTC": Incident detected
    - "14:32 UTC": Incident declared, team assembled
    - "14:35 UTC": Investigation started
    - "14:45 UTC": Root cause identified
    - "14:50 UTC": Rollback initiated
    - "14:55 UTC": Service restored
    - "15:15 UTC": Incident resolved

  action_items:
    - owner: "team-lead"
      task: "Improve deployment validation"
      due_date: "2024-01-22"

    - owner: "senior-engineer"
      task: "Add pre-deployment smoke tests"
      due_date: "2024-01-22"

  lessons_learned:
    - "Deployment validation caught the error but was ignored"
    - "Smoke tests would have caught the issue immediately"
    - "Communication was effective and timely"
  ```

exercises:
  - Write postmortem for a sample incident
  - Conduct root cause analysis
  - Create action items
  - Present findings to team
```

### Module 4: Performance Management

#### Learning Objectives
- Understand performance metrics
- Monitor system performance
- Identify performance bottlenecks
- Implement performance optimizations

#### Session 1: Performance Monitoring

```yaml
duration: 2 hours
format: Workshop

topics:
  - Key Performance Indicators (KPIs)
  - Performance Metrics:
    - Response Times (P50, P95, P99)
    - Throughput
    - Error Rate
    - Cache Hit Rate
  - Performance Trends
  - Anomaly Detection

hands-on:
  - Monitor performance metrics
  ```bash
  # View real-time metrics
  wrangler tail --env production --format=pretty

  # Export performance data
  wrangler analytics export --start 2024-01-01 --end 2024-01-31 --format csv > metrics.csv
  ```

  - Analyze performance trends
  ```javascript
  // Performance analysis script
  const performanceAnalysis = {
    timeframe: '7d',
    metrics: {
      responseTime: {
        p50: 45,
        p95: 180,
        p99: 420,
        trend: 'stable'
      },
      throughput: {
        requestsPerSecond: 1250,
        trend: 'increasing'
      },
      errorRate: {
        value: 0.002,
        threshold: 0.01,
        trend: 'decreasing'
      },
      cacheHitRate: {
        value: 0.87,
        threshold: 0.80,
        trend: 'stable'
      }
    }
  };
  ```

exercises:
  - Monitor performance metrics in real-time
  - Analyze performance trends
  - Identify performance anomalies
  - Create performance reports
```

#### Session 2: Performance Optimization

```yaml
duration: 3 hours
format: Workshop

topics:
  - Bottleneck Identification
  - Optimization Strategies:
    - Caching Optimization
    - Query Optimization
    - Resource Scaling
    - Configuration Tuning
  - Load Testing
  - Performance Benchmarking

hands-on:
  - Run load test
  ```javascript
  // Load testing script
  const loadTestConfig = {
    url: 'https://api.yourdomain.com/chat',
    maxConcurrent: 100,
    duration: 60000, // 1 minute
    rampUp: 10000 // 10 seconds
  };

  async function runLoadTest(config) {
    const results = [];
    const startTime = Date.now();

    while (Date.now() - startTime < config.duration) {
      const promises = [];

      for (let i = 0; i < config.maxConcurrent; i++) {
        promises.push(
          fetch(config.url)
            .then(res => ({
              success: res.ok,
              status: res.status,
              startTime: performance.now(),
              endTime: performance.now()
            }))
        );
      }

      const batchResults = await Promise.all(promises);
      results.push(...batchResults);
    }

    return analyzeResults(results);
  }

  function analyzeResults(results) {
    const successful = results.filter(r => r.success);
    const durations = successful.map(r => r.endTime - r.startTime);

    return {
      totalRequests: results.length,
      successfulRequests: successful.length,
      failedRequests: results.length - successful.length,
      successRate: (successful.length / results.length * 100).toFixed(2) + '%',
      avgDuration: durations.reduce((a, b) => a + b, 0) / durations.length,
      minDuration: Math.min(...durations),
      maxDuration: Math.max(...durations),
      p95Duration: durations.sort((a, b) => a - b)[Math.floor(durations.length * 0.95)]
    };
  }
  ```

  - Optimize configuration
  ```javascript
  // Configuration optimization
  const optimizedConfig = {
    cache: {
      size: 200, // Increased from 100
      ttl: 7200000, // Increased from 3600000 (2 hours)
      cleanupInterval: 300000 // 5 minutes
    },
    rateLimiting: {
      maxRequests: 2000, // Increased from 1000
      windowSize: 60000,
      burstSize: 50 // Increased from 20
    },
    concurrency: {
      max: 100, // Increased from 50
      min: 10
    }
  };
  ```

exercises:
  - Identify performance bottlenecks
  - Implement caching optimizations
  - Tune rate limiting parameters
  - Measure optimization impact
```

### Module 5: Maintenance Procedures

#### Learning Objectives
- Understand maintenance schedules
- Master backup procedures
- Learn about system updates
- Handle capacity planning

#### Session 1: Backup and Recovery

```yaml
duration: 2 hours
format: Workshop

topics:
  - Data Backup Strategies
  - Backup Automation
  - Recovery Procedures
  - Disaster Recovery Planning

hands-on:
  - Create backup automation
  ```bash
  #!/bin/bash
  # backup-worker-config.sh

  BACKUP_DIR="/backups/workers"
  DATE=$(date +%Y%m%d_%H%M%S)

  # Create backup directory
  mkdir -p $BACKUP_DIR/$DATE

  # Backup configuration
  wrangler secret list --env production > $BACKUP_DIR/$DATE/secrets.txt

  # Backup KV data
  wrangler kv:bulk export --namespace-id=<namespace-id> --path=$BACKUP_DIR/$DATE/kv-export.json

  # Backup metrics
  wrangler analytics export --start $(date -d '30 days ago' +%Y-%m-%d) --end $(date +%Y-%m-%d) --path=$BACKUP_DIR/$DATE/metrics.csv

  # Compress backup
  tar -czf $BACKUP_DIR/backup-$DATE.tar.gz $BACKUP_DIR/$DATE

  # Clean up old backups (keep last 30 days)
  find $BACKUP_DIR -name "backup-*.tar.gz" -mtime +30 -delete

  echo "Backup completed: $BACKUP_DIR/backup-$DATE.tar.gz"
  ```

  - Test recovery procedure
  ```bash
  #!/bin/bash
  # restore-worker-config.sh

  BACKUP_FILE=$1

  if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file>"
    exit 1
  fi

  # Extract backup
  tar -xzf $BACKUP_FILE -C /tmp

  # Restore configuration
  echo "Restoring configuration..."

  # Restore KV data
  wrangler kv:bulk import --namespace-id=<namespace-id> --path=/tmp/kv-export.json

  echo "Restore completed"
  ```

exercises:
  - Automate daily backups
  - Test backup and recovery procedures
  - Create disaster recovery plan
  - Document recovery procedures
```

#### Session 2: System Updates

```yaml
duration: 2 hours
format: Workshop

topics:
  - Update Strategies
  - Update Testing
  - Deployment Procedures
  - Rollback Plans

hands-on:
  - Create update procedure
  ```bash
  #!/bin/bash
  # update-worker.sh

  VERSION=$1
  ENVIRONMENT=$2

  if [ -z "$VERSION" ] || [ -z "$ENVIRONMENT" ]; then
    echo "Usage: $0 <version> <environment>"
    exit 1
  }

  echo "Updating worker to version $VERSION in $ENVIRONMENT environment"

  # Create backup
  ./backup-worker-config.sh

  # Deploy update
  wrangler deploy --env $ENVIRONMENT --version $VERSION

  # Run smoke tests
  npm run test:smoke

  # Check health
  curl -f https://api.yourdomain.com/health || {
    echo "Health check failed, rolling back..."
    wrangler rollback --env $ENVIRONMENT
    exit 1
  }

  echo "Update completed successfully"
  ```

exercises:
  - Create update procedures
  - Test update in staging
  - Plan and execute production update
  - Document update procedures
```

## 📊 Certification Assessment

### Written Exam

```yaml
duration: 1 hour
format: Multiple choice + short answer

topics:
  - Platform architecture
  - Deployment procedures
  - Monitoring and alerting
  - Incident response
  - Performance optimization

passing_score: 80%
```

### Practical Exam

```yaml
duration: 3 hours
format: Hands-on tasks

tasks:
  1. Deploy a worker to production
  2. Configure monitoring and alerting
  3. Respond to a simulated incident
  4. Optimize performance

evaluation_criteria:
  - Task completion
  - Process correctness
  - Documentation quality
  - Time management
  - Best practices adherence
```

## 📖 Additional Resources

### Documentation
- [Deployment Guide](../operations/deployment.md)
- [Monitoring Setup](../operations/monitoring.md)
- [Performance Tuning](../operations/performance-tuning.md)
- [Incident Response](../operations/incident-response.md)

### Tools
- [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/)
- [Cloudflare Dashboard](https://dash.cloudflare.com/)
- [Grafana](https://grafana.com/)
- [PagerDuty](https://www.pagerduty.com/)

### Best Practices
- [Site Reliability Engineering](https://sre.google/)
- [Incident Management](https://www.atlassian.com/incident-management)
- [Monitoring Best Practices](https://www.datadoghq.com/blog/monitoring-101/)
- [DevOps Practices](https://www.atlassian.com/devops)

## 🔗 Related Documentation

- [Developer Training](./developer-training.md)
- [Performance Workshop](./performance-workshop.md)
- [Deployment Guide](../operations/deployment.md)
- [Incident Response](../operations/incident-response.md)
