# Incident Response

## 🚨 Incident Response Overview

This guide covers the complete incident response process for enhanced Cloudflare Workers, including incident classification, response procedures, communication protocols, and post-incident analysis. Effective incident response minimizes downtime and prevents recurrence.

## 📋 Incident Classification

### Severity Levels

#### P0 - Critical
```
Definition: Service completely down or severely degraded
Examples:
  - Complete service outage
  - Data loss or corruption
  - Security breach
  - Revenue loss >$10,000/hour

Response Time: <15 minutes
Escalation: Immediately to CEO and VP Engineering
```

#### P1 - High
```
Definition: Major functionality impaired
Examples:
  - Significant performance degradation
  - Partial service outage
  - Feature unavailable for >50% of users
  - Revenue loss >$1,000/hour

Response Time: <30 minutes
Escalation: VP Engineering and Team Lead
```

#### P2 - Medium
```
Definition: Minor functionality issues
Examples:
  - Slow response times
  - Feature unavailable for <50% of users
  - Reduced quality of service
  - Revenue loss <$1,000/hour

Response Time: <1 hour
Escalation: Team Lead and Senior Engineer
```

#### P3 - Low
```
Definition: Cosmetic or minor issues
Examples:
  - Occasional errors
  - UI glitches
  - Documentation issues
  - No revenue impact

Response Time: <4 hours
Escalation: Team Lead
```

### Incident Categories

```yaml
technical_incidents:
  - Service Outage
  - Performance Degradation
  - Database Failure
  - Network Issues
  - API Failures

security_incidents:
  - Unauthorized Access
  - Data Breach
  - DDoS Attack
  - Malware Infection
  - Vulnerability Exploitation

operational_incidents:
  - Deployment Failure
  - Configuration Error
  - Resource Exhaustion
  - Backup Failure
  - Human Error
```

## 🚀 Incident Response Process

### 1. Detection

```javascript
// ai-proxy/src/incident-detector.js

export class IncidentDetector {
  constructor(config) {
    this.config = config;
    this.activeIncidents = new Map();
    this.alertThresholds = config.alertThresholds || {
      errorRate: 0.05, // 5% error rate
      responseTimeP95: 2000, // 2 seconds
      availability: 99.5 // 99.5% uptime
    };
  }

  detectIncident(metrics) {
    const incidents = [];

    // Check error rate
    if (metrics.errorRate > this.alertThresholds.errorRate) {
      incidents.push({
        type: 'error_rate',
        severity: this.calculateSeverity('error_rate', metrics.errorRate),
        value: metrics.errorRate,
        threshold: this.alertThresholds.errorRate,
        impact: this.estimateImpact('error_rate', metrics.errorRate)
      });
    }

    // Check response time
    if (metrics.responseTimeP95 > this.alertThresholds.responseTimeP95) {
      incidents.push({
        type: 'response_time',
        severity: this.calculateSeverity('response_time', metrics.responseTimeP95),
        value: metrics.responseTimeP95,
        threshold: this.alertThresholds.responseTimeP95,
        impact: this.estimateImpact('response_time', metrics.responseTimeP95)
      });
    }

    // Check availability
    if (metrics.availability < this.alertThresholds.availability) {
      incidents.push({
        type: 'availability',
        severity: this.calculateSeverity('availability', metrics.availability),
        value: metrics.availability,
        threshold: this.alertThresholds.availability,
        impact: this.estimateImpact('availability', metrics.availability)
      });
    }

    return incidents;
  }

  calculateSeverity(type, value) {
    const thresholds = {
      error_rate: { critical: 0.1, high: 0.05, medium: 0.02 },
      response_time: { critical: 5000, high: 2000, medium: 1000 },
      availability: { critical: 99, high: 99.5, medium: 99.9 }
    };

    const typeThresholds = thresholds[type] || {};
    
    if (type === 'availability') {
      if (value < typeThresholds.critical) return 'P0';
      if (value < typeThresholds.high) return 'P1';
      if (value < typeThresholds.medium) return 'P2';
    } else {
      if (value > typeThresholds.critical) return 'P0';
      if (value > typeThresholds.high) return 'P1';
      if (value > typeThresholds.medium) return 'P2';
    }

    return 'P3';
  }

  estimateImpact(type, value) {
    // Estimate business impact
    const impactModels = {
      error_rate: (rate) => ({
        usersAffected: Math.min(100, rate * 1000),
        revenueImpact: rate * 100,
        severity: rate > 0.1 ? 'critical' : 'high'
      }),
      response_time: (time) => ({
        usersAffected: Math.min(80, time / 50),
        revenueImpact: time / 100,
        severity: time > 5000 ? 'critical' : 'high'
      }),
      availability: (avail) => ({
        usersAffected: (100 - avail) * 10,
        revenueImpact: (100 - avail) * 10,
        severity: avail < 99 ? 'critical' : 'high'
      })
    };

    return impactModels[type] ? impactModels[type](value) : { usersAffected: 0, revenueImpact: 0 };
  }
}
```

### 2. Triage

```javascript
// ai-proxy/src/incident-triage.js

export class IncidentTriage {
  constructor() {
    this.incidentHistory = [];
  }

  async triageIncident(incident) {
    const triage = {
      incidentId: this.generateIncidentId(),
      timestamp: new Date().toISOString(),
      severity: incident.severity,
      category: this.classifyCategory(incident),
      priority: this.calculatePriority(incident),
      assignee: this.assignIncident(incident),
      estimatedResolutionTime: this.estimateResolutionTime(incident),
      actions: this.recommendActions(incident),
      similarIncidents: this.findSimilarIncidents(incident)
    };

    this.incidentHistory.push(triage);
    return triage;
  }

  generateIncidentId() {
    const timestamp = Date.now().toString(36).toUpperCase();
    const random = Math.random().toString(36).substring(2, 6).toUpperCase();
    return `INC-${timestamp}-${random}`;
  }

  classifyCategory(incident) {
    const categories = {
      'error_rate': 'technical',
      'response_time': 'technical',
      'availability': 'technical',
      'security': 'security',
      'deployment': 'operational'
    };

    return categories[incident.type] || 'operational';
  }

  calculatePriority(incident) {
    const severityScores = { 'P0': 1, 'P1': 2, 'P2': 3, 'P3': 4 };
    const impactScores = { 'critical': 1, 'high': 2, 'medium': 3, 'low': 4 };

    return severityScores[incident.severity] * impactScores[incident.impact.severity];
  }

  assignIncident(incident) {
    // Assign based on severity and expertise
    const assignmentRules = {
      'P0': 'team-lead',
      'P1': 'senior-engineer',
      'P2': 'engineer',
      'P3': 'junior-engineer'
    };

    return assignmentRules[incident.severity] || 'engineer';
  }

  estimateResolutionTime(incident) {
    const resolutionTimes = {
      'P0': '1 hour',
      'P1': '2 hours',
      'P2': '4 hours',
      'P3': '8 hours'
    };

    return resolutionTimes[incident.severity] || '8 hours';
  }

  recommendActions(incident) {
    const actionTemplates = {
      'error_rate': [
        'Check error logs',
        'Identify common error patterns',
        'Roll back recent changes if needed',
        'Scale up resources'
      ],
      'response_time': [
        'Analyze performance metrics',
        'Check database queries',
        'Review recent code changes',
        'Optimize slow operations'
      ],
      'availability': [
        'Verify service status',
        'Check dependencies',
        'Review recent deployments',
        'Restore from backup if needed'
      ]
    };

    return actionTemplates[incident.type] || [
      'Gather more information',
      'Analyze logs',
      'Consult with team',
      'Document findings'
    ];
  }

  findSimilarIncidents(incident) {
    // Find similar historical incidents
    return this.incidentHistory
      .filter(h => h.category === this.classifyCategory(incident))
      .slice(-5);
  }
}
```

### 3. Response

```javascript
// ai-proxy/src/incident-responder.js

export class IncidentResponder {
  constructor(context) {
    this.context = context;
    this.activeResponse = null;
  }

  async startResponse(incident) {
    this.activeResponse = {
      incidentId: incident.incidentId,
      startTime: Date.now(),
      responders: [],
      actions: [],
      status: 'responding',
      timeline: []
    };

    // Add initial actions
    await this.executeAction('acknowledge', {
      message: `Incident ${incident.incidentId} acknowledged`,
      timestamp: new Date().toISOString()
    });

    // Notify responders
    await this.notifyTeam(incident);

    // Begin investigation
    await this.beginInvestigation(incident);

    return this.activeResponse;
  }

  async executeAction(actionType, details) {
    const action = {
      type: actionType,
      timestamp: new Date().toISOString(),
      ...details
    };

    this.activeResponse.actions.push(action);
    this.activeResponse.timeline.push({
      timestamp: action.timestamp,
      action: actionType,
      details
    });

    return action;
  }

  async notifyTeam(incident) {
    const notification = {
      incidentId: incident.incidentId,
      severity: incident.severity,
      category: incident.category,
      message: `Incident ${incident.severity}: ${incident.type}`,
      assignee: incident.assignee
    };

    // Send notifications to configured channels
    await this.sendSlackNotification(notification);
    await this.sendPagerDutyNotification(notification);
    await this.sendEmailNotification(notification);
  }

  async beginInvestigation(incident) {
    await this.executeAction('investigation_started', {
      message: `Investigation started for ${incident.incidentId}`,
      steps: incident.actions
    });

    // Gather diagnostic information
    const diagnostics = await this.gatherDiagnostics(incident);
    await this.executeAction('diagnostics_collected', diagnostics);
  }

  async gatherDiagnostics(incident) {
    return {
      metrics: await this.collectMetrics(),
      logs: await this.collectLogs(),
      traces: await this.collectTraces(),
      systemStatus: await this.checkSystemStatus()
    };
  }

  async collectMetrics() {
    return {
      timestamp: new Date().toISOString(),
      requestRate: this.context.metrics.requestCount / 60,
      errorRate: this.context.metrics.errorRate,
      responseTimeP95: this.context.metrics.responseTimeP95,
      cacheHitRate: this.context.metrics.cacheHitRate,
      memoryUsage: this.context.memoryProfiler.getCurrentUsage()
    };
  }

  async collectLogs() {
    // Collect recent error logs
    return {
      timestamp: new Date().toISOString(),
      logCount: 100,
      sampleLogs: [
        'Error: Timeout while connecting to API',
        'Warning: High memory usage detected',
        'Info: Cache miss for request'
      ]
    };
  }

  async collectTraces() {
    // Collect request traces
    return {
      timestamp: new Date().toISOString(),
      traceCount: 50,
      slowestTraces: [
        { requestId: '123', duration: 2500, endpoint: '/api/chat' },
        { requestId: '456', duration: 1800, endpoint: '/api/analyze' }
      ]
    };
  }

  async checkSystemStatus() {
    return {
      cache: 'operational',
      providers: 'degraded',
      database: 'operational',
      loadBalancer: 'operational'
    };
  }

  async implementMitigation(mitigation) {
    await this.executeAction('mitigation_attempted', {
      message: `Attempting mitigation: ${mitigation.type}`,
      mitigation: mitigation
    });

    // Execute mitigation
    const result = await this.executeMitigation(mitigation);
    
    await this.executeAction('mitigation_completed', {
      message: `Mitigation ${result.success ? 'succeeded' : 'failed'}`,
      result
    });

    return result;
  }

  async executeMitigation(mitigation) {
    switch (mitigation.type) {
      case 'rollback':
        return await this.rollbackDeployment(mitigation.version);
      case 'scale_up':
        return await this.scaleResources(mitigation.target);
      case 'clear_cache':
        return await this.clearCache();
      case 'restart_service':
        return await this.restartService();
      default:
        return { success: false, message: 'Unknown mitigation type' };
    }
  }

  async rollbackDeployment(version) {
    // Implement rollback logic
    return { success: true, message: `Rolled back to version ${version}` };
  }

  async scaleResources(target) {
    // Implement scaling logic
    return { success: true, message: `Scaled resources to ${target}` };
  }

  async clearCache() {
    // Implement cache clearing
    return { success: true, message: 'Cache cleared' };
  }

  async restartService() {
    // Implement service restart
    return { success: true, message: 'Service restarted' };
  }

  async completeResponse(resolution) {
    this.activeResponse.status = 'resolved';
    this.activeResponse.endTime = Date.now();
    this.activeResponse.resolution = resolution;
    this.activeResponse.duration = this.activeResponse.endTime - this.activeResponse.startTime;

    await this.executeAction('incident_resolved', {
      message: `Incident ${this.activeResponse.incidentId} resolved`,
      resolution
    });

    return this.activeResponse;
  }
}
```

## 📢 Communication Protocol

### 1. Internal Communication

```javascript
// ai-proxy/src/communication.js

export class IncidentCommunicator {
  constructor(config) {
    this.config = config;
    this.channels = {
      slack: config.slackWebhook,
      pagerDuty: config.pagerDutyKey,
      email: config.emailRecipients
    };
  }

  async notifyTeam(incident, message) {
    const notifications = [
      this.sendSlackMessage(incident, message),
      this.sendPagerDutyAlert(incident, message),
      this.sendEmailNotification(incident, message)
    ];

    await Promise.all(notifications);
  }

  async sendSlackMessage(incident, message) {
    const payload = {
      channel: this.config.slackChannel,
      attachments: [{
        color: this.getSeverityColor(incident.severity),
        title: `Incident ${incident.severity}: ${incident.incidentId}`,
        text: message,
        fields: [
          { title: 'Severity', value: incident.severity, short: true },
          { title: 'Category', value: incident.category, short: true },
          { title: 'Impact', value: incident.impact.severity, short: true },
          { title: 'Assignee', value: incident.assignee, short: true }
        ],
        timestamp: new Date().toISOString()
      }]
    };

    await fetch(this.channels.slack, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
  }

  async sendPagerDutyAlert(incident, message) {
    const payload = {
      routing_key: this.channels.pagerDuty,
      event_action: 'trigger',
      payload: {
        summary: `Incident ${incident.severity}: ${incident.incidentId}`,
        severity: incident.severity.toLowerCase(),
        source: 'cloudflare-workers',
        custom_details: {
          category: incident.category,
          impact: incident.impact
        }
      }
    };

    await fetch('https://events.pagerduty.com/v2/enqueue', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
  }

  async sendEmailNotification(incident, message) {
    const payload = {
      to: this.channels.email,
      subject: `Incident ${incident.severity}: ${incident.incidentId}`,
      text: message,
      html: this.formatEmailMessage(incident, message)
    };

    // Send email via your email service
    await fetch('/api/send-email', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
  }

  async sendStatusUpdate(incident, update) {
    const message = `
      Incident Update: ${incident.incidentId}
      Status: ${update.status}
      Time: ${new Date().toISOString()}
      Details: ${update.message}
    `;

    await this.sendSlackMessage(incident, message);
  }

  getSeverityColor(severity) {
    const colors = {
      'P0': 'danger',
      'P1': 'warning',
      'P2': 'good',
      'P3': '#439FE0'
    };

    return colors[severity] || 'good';
  }

  formatEmailMessage(incident, message) {
    return `
      <h2>Incident ${incident.severity}: ${incident.incidentId}</h2>
      <p><strong>Category:</strong> ${incident.category}</p>
      <p><strong>Severity:</strong> ${incident.severity}</p>
      <p><strong>Status:</strong> ${incident.status}</p>
      <hr>
      <p>${message}</p>
    `;
  }
}
```

### 2. External Communication

```javascript
// ai-proxy/src/external-communication.js

export class ExternalCommunicator {
  constructor(config) {
    this.config = config;
    this.statusPageUrl = config.statusPageUrl;
  }

  async updateStatusPage(incident) {
    const statusUpdate = {
      incident_id: incident.incidentId,
      name: incident.description || 'Service Issue',
      status: this.mapStatusToStatusPage(incident.status),
      impact: this.mapSeverityToImpact(incident.severity),
      message: this.formatStatusMessage(incident),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    await fetch(`${this.statusPageUrl}/api/incidents`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(statusUpdate)
    });
  }

  async publishIncidentPostmortem(postmortem) {
    const postmortemUrl = `${this.statusPageUrl}/incidents/${postmortem.incidentId}/postmortem`;
    
    await fetch(postmortemUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: postmortem.title,
        summary: postmortem.summary,
        rootCause: postmortem.rootCause,
        timeline: postmortem.timeline,
        resolution: postmortem.resolution,
        actionItems: postmortem.actionItems,
        publishedAt: new Date().toISOString()
      })
    });
  }

  mapStatusToStatusPage(status) {
    const mapping = {
      'investigating': 'investigating',
      'identified': 'identified',
      'monitoring': 'monitoring',
      'resolved': 'resolved'
    };

    return mapping[status] || 'investigating';
  }

  mapSeverityToImpact(severity) {
    const mapping = {
      'P0': 'critical',
      'P1': 'major',
      'P2': 'minor',
      'P3': 'maintenance'
    };

    return mapping[severity] || 'none';
  }

  formatStatusMessage(incident) {
    return `We are currently experiencing issues with ${incident.category}. ${incident.severity === 'P0' ? 'Our team is actively working on a fix.' : 'We are investigating the issue and will provide updates as soon as possible.'}`;
  }
}
```

## 📝 Post-Incident Analysis

### 1. Postmortem Creation

```javascript
// ai-proxy/src/postmortem.js

export class PostmortemGenerator {
  constructor() {
    this.template = {
      incidentId: '',
      title: '',
      severity: '',
      date: '',
      summary: '',
      impact: '',
      rootCause: '',
      timeline: [],
      resolution: '',
      actionItems: [],
      lessonsLearned: [],
      acknowledgments: []
    };
  }

  generatePostmortem(incident, response) {
    return {
      ...this.template,
      incidentId: incident.incidentId,
      title: this.generateTitle(incident),
      severity: incident.severity,
      date: new Date().toISOString(),
      summary: this.generateSummary(incident, response),
      impact: this.assessImpact(incident, response),
      rootCause: this.identifyRootCause(response),
      timeline: this.buildTimeline(response),
      resolution: this.documentResolution(response),
      actionItems: this.generateActionItems(response),
      lessonsLearned: this.extractLessons(response),
      acknowledgments: this.generateAcknowledgments(response)
    };
  }

  generateTitle(incident) {
    return `${incident.severity} ${incident.category} Incident - ${new Date().toLocaleDateString()}`;
  }

  generateSummary(incident, response) {
    return `On ${new Date().toLocaleDateString()}, we experienced a ${incident.severity} ${incident.category} incident that affected ${incident.impact.usersAffected}% of users. The incident was detected at ${response.startTime} and resolved at ${response.endTime}.`;
  }

  assessImpact(incident, response) {
    return {
      usersAffected: incident.impact.usersAffected,
      duration: response.duration,
      revenueImpact: incident.impact.revenueImpact,
      severity: incident.impact.severity
    };
  }

  identifyRootCause(response) {
    // Analyze response data to identify root cause
    const diagnostics = response.actions.find(a => a.type === 'diagnostics_collected');
    return diagnostics?.rootCause || 'Under investigation';
  }

  buildTimeline(response) {
    return response.timeline.map(item => ({
      time: item.timestamp,
      event: item.action,
      details: item.details
    }));
  }

  documentResolution(response) {
    return response.resolution?.description || 'Incident resolved through standard procedures';
  }

  generateActionItems(response) {
    // Generate action items based on incident learnings
    return [
      { owner: 'team-lead', task: 'Review incident response procedures', dueDate: '1 week' },
      { owner: 'senior-engineer', task: 'Implement monitoring improvements', dueDate: '2 weeks' },
      { owner: 'engineer', task: 'Update runbooks', dueDate: '1 week' }
    ];
  }

  extractLessons(response) {
    return [
      'Detection systems functioned correctly',
      'Response team responded quickly',
      'Communication channels were effective',
      'Mitigation procedures need refinement'
    ];
  }

  generateAcknowledgments(response) {
    return response.responders.map(responder => responder.name);
  }
}
```

### 2. Incident Review Meeting

```yaml
incident_review_meeting:
  attendees:
    - Incident responders
    - Team lead
    - Engineering manager
    - Customer support (if applicable)
  
  agenda:
    - Overview of incident (5 min)
    - Timeline review (10 min)
    - Root cause analysis (15 min)
    - Response effectiveness (10 min)
    - Action items discussion (10 min)
    - Lessons learned (10 min)
    - Next steps (5 min)
  
  outcomes:
    - Finalized postmortem document
    - Assigned action items with due dates
    - Updated runbooks and procedures
    - Scheduled follow-up reviews
```

### 3. Continuous Improvement

```javascript
// ai-proxy/src/improvement-tracker.js

export class ImprovementTracker {
  constructor() {
    this.improvements = [];
  }

  trackImprovement(incidentId, improvement) {
    this.improvements.push({
      incidentId,
      improvement,
      status: 'pending',
      createdAt: new Date().toISOString(),
      completedAt: null
    });
  }

  async implementImprovement(improvementId) {
    const improvement = this.improvements.find(i => i.id === improvementId);
    if (!improvement) throw new Error('Improvement not found');

    improvement.status = 'in-progress';

    try {
      await this.executeImprovement(improvement);
      improvement.status = 'completed';
      improvement.completedAt = new Date().toISOString();
    } catch (error) {
      improvement.status = 'failed';
      improvement.error = error.message;
    }

    return improvement;
  }

  async executeImprovement(improvement) {
    switch (improvement.type) {
      case 'monitoring':
        return await this.improveMonitoring(improvement);
      case 'alerting':
        return await this.improveAlerting(improvement);
      case 'procedure':
        return await this.updateProcedures(improvement);
      case 'documentation':
        return await this.updateDocumentation(improvement);
      default:
        throw new Error('Unknown improvement type');
    }
  }

  getOpenImprovements() {
    return this.improvements.filter(i => i.status !== 'completed');
  }

  getImprovementsByIncident(incidentId) {
    return this.improvements.filter(i => i.incidentId === incidentId);
  }
}
```

## 🛠️ Runbooks

### Common Incident Runbooks

#### 1. High Error Rate
```yaml
title: High Error Rate
severity: P1
owner: Senior Engineer

steps:
  1. Check error logs for common patterns
  2. Identify affected endpoints and users
  3. Check recent code changes and deployments
  4. Roll back recent changes if needed
  5. Check database connectivity and performance
  6. Verify third-party API status
  7. Scale resources if necessary
  8. Monitor error rate until it returns to normal

escalation:
  - If not resolved in 30 minutes, escalate to Team Lead
  - If not resolved in 1 hour, escalate to VP Engineering
```

#### 2. Slow Response Times
```yaml
title: Slow Response Times
severity: P2
owner: Engineer

steps:
  1. Analyze performance metrics and identify slowest operations
  2. Check database query performance
  3. Review cache hit rate and effectiveness
  4. Check for network latency issues
  5. Profile code to identify bottlenecks
  6. Optimize slow queries and operations
  7. Scale resources if necessary
  8. Monitor response times until they return to normal

escalation:
  - If response time >5 seconds, escalate to Senior Engineer
  - If not resolved in 2 hours, escalate to Team Lead
```

#### 3. Service Outage
```yaml
title: Service Outage
severity: P0
owner: Team Lead

steps:
  1. Immediately declare incident and notify team
  2. Check service health endpoints
  3. Identify scope of outage (affected services/users)
  4. Attempt immediate rollback of recent changes
  5. Check dependencies and third-party services
  6. Implement emergency fix or workaround
  7. Scale resources if necessary
  8. Monitor service until fully restored
  9. Conduct post-incident review

escalation:
  - Immediately escalate to VP Engineering
  - Notify stakeholders and customers
  - Consider incident command if outage >30 minutes
```

## 📊 Incident Metrics

### Key Performance Indicators

```yaml
incident_metrics:
  mttr: "Mean Time To Resolution"
  mtti: "Mean Time To Identification"
  mttm: "Mean Time To Mitigation"
  incident_frequency: "Number of incidents per month"
  repeat_incidents: "Incidents that occurred more than once"
  
target_values:
  mttr: "<1 hour for P1 incidents"
  mtti: "<15 minutes for P0 incidents"
  mttm: "<30 minutes for P0 incidents"
  incident_frequency: "<5 incidents per month"
  repeat_incidents: "<10% of total incidents"
```

### Metric Tracking

```javascript
// ai-proxy/src/metrics.js

export class IncidentMetrics {
  constructor() {
    this.incidents = [];
  }

  recordIncident(incident) {
    this.incidents.push({
      ...incident,
      recordedAt: new Date().toISOString()
    });
  }

  calculateMTTR() {
    const resolvedIncidents = this.incidents.filter(i => i.endTime);
    if (resolvedIncidents.length === 0) return null;

    const totalResolutionTime = resolvedIncidents.reduce(
      (sum, i) => sum + (i.endTime - i.startTime),
      0
    );

    return totalResolutionTime / resolvedIncidents.length;
  }

  calculateMTTI() {
    const identifiedIncidents = this.incidents.filter(i => i.identificationTime);
    if (identifiedIncidents.length === 0) return null;

    const totalIdentificationTime = identifiedIncidents.reduce(
      (sum, i) => sum + (i.identificationTime - i.startTime),
      0
    );

    return totalIdentificationTime / identifiedIncidents.length;
  }

  calculateMTTM() {
    const mitigatedIncidents = this.incidents.filter(i => i.mitigationTime);
    if (mitigatedIncidents.length === 0) return null;

    const totalMitigationTime = mitigatedIncidents.reduce(
      (sum, i) => sum + (i.mitigationTime - i.startTime),
      0
    );

    return totalMitigationTime / mitigatedIncidents.length;
  }

  getIncidentFrequency(timeframe = 'month') {
    const now = Date.now();
    const timeframes = {
      'week': 7 * 24 * 60 * 60 * 1000,
      'month': 30 * 24 * 60 * 60 * 1000,
      'year': 365 * 24 * 60 * 60 * 1000
    };

    const cutoff = now - (timeframes[timeframe] || timeframes['month']);

    return this.incidents.filter(i => new Date(i.startTime).getTime() > cutoff).length;
  }
}
```

## 🔗 Related Documentation

- [Monitoring Setup](./monitoring.md)
- [Deployment Guide](./deployment.md)
- [Performance Tuning](./performance-tuning.md)
- [Troubleshooting Guide](../user-guides/troubleshooting.md)
- [Architecture Overview](../developer-docs/architecture.md)
