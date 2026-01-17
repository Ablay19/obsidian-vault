# Troubleshooting Guide

This guide provides solutions for common issues encountered in the AI Platform deployment and operation.

## 🔍 General Debugging

### Check System Status

```bash
# Check Kubernetes pods
kubectl get pods -n arch-separation
kubectl get pods -n monitoring

# Check pod logs
kubectl logs -f deployment/api-gateway -n arch-separation
kubectl logs -f deployment/prometheus -n monitoring

# Check ingress status
kubectl get ingress -n arch-separation
kubectl describe ingress api-gateway -n arch-separation
```

### Health Checks

```bash
# API Gateway health
curl https://api.your-domain.com/health

# Worker health
curl https://ai-worker.your-domain.workers.dev/health

# Monitoring health
curl http://localhost:9090/-/healthy
```

## 🚨 Common Issues and Solutions

### 1. API Gateway Issues

#### High Latency or Timeouts

**Symptoms**:
- API requests taking >5 seconds
- 504 Gateway Timeout errors

**Diagnosis**:
```bash
# Check pod resource usage
kubectl top pods -n arch-separation

# Check database connections
kubectl exec -it deployment/api-gateway -n arch-separation -- netstat -tlnp | grep :5432

# Check circuit breaker status
curl https://api.your-domain.com/health | jq .checks.circuit_breaker
```

**Solutions**:
1. **Scale horizontally**: `kubectl scale deployment api-gateway --replicas=3 -n arch-separation`
2. **Check database performance**: Monitor slow queries
3. **Reset circuit breaker**: Restart pod if stuck open
4. **Check network policies**: Ensure proper inter-service communication

#### 5xx Errors

**Symptoms**:
- 500 Internal Server Error
- 502 Bad Gateway

**Diagnosis**:
```bash
# Check application logs
kubectl logs -f deployment/api-gateway -n arch-separation --tail=100

# Check dependencies
kubectl get pods -n arch-separation | grep -v Running

# Verify configuration
kubectl describe configmap api-gateway-config -n arch-separation
```

**Solutions**:
1. **Check database connectivity**: Verify DB credentials and network
2. **Validate environment variables**: Ensure all required vars are set
3. **Review recent deployments**: Check for breaking changes
4. **Check memory/CPU limits**: Pods may be OOM killed

### 2. Cloudflare Workers Issues

#### Worker Not Responding

**Symptoms**:
- 522 Connection Timed Out
- Worker requests failing

**Diagnosis**:
```bash
# Check worker logs
wrangler tail

# Verify KV namespace
wrangler kv:key list --namespace-id <namespace-id>

# Check R2 bucket
wrangler r2 object list ai-platform-videos
```

**Solutions**:
1. **Redeploy worker**: `wrangler deploy --env production`
2. **Check environment variables**: Verify API_GATEWAY_URL and tokens
3. **Validate KV/R2 access**: Ensure proper permissions
4. **Check rate limits**: Monitor Cloudflare dashboard

#### Slow Worker Performance

**Symptoms**:
- Response times >2 seconds
- High CPU time in logs

**Diagnosis**:
```bash
# Check worker analytics
wrangler tail --format=json | jq '.cpuTime'

# Monitor KV operations
wrangler kv:key list --namespace-id <namespace-id> | wc -l
```

**Solutions**:
1. **Optimize code**: Reduce bundle size, improve algorithms
2. **Cache frequently accessed data**: Use in-memory caching
3. **Batch KV operations**: Reduce individual KV calls
4. **Check for memory leaks**: Monitor heap usage

### 3. Database Issues

#### Connection Pool Exhaustion

**Symptoms**:
- Database connection errors
- Slow query responses

**Diagnosis**:
```bash
# Check connection count
kubectl exec -it deployment/postgres -- psql -c "SELECT count(*) FROM pg_stat_activity;"

# Monitor connection pool
kubectl exec -it deployment/api-gateway -- netstat -tlnp | grep :5432 | wc -l
```

**Solutions**:
1. **Increase connection pool size**: Update database configuration
2. **Optimize queries**: Add indexes, rewrite slow queries
3. **Implement connection pooling**: Use pgxpool or similar
4. **Scale database**: Add read replicas

#### Slow Queries

**Symptoms**:
- API responses delayed
- Database CPU high

**Diagnosis**:
```sql
-- Find slow queries
SELECT query, total_exec_time, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Check for missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE schemaname = 'public' AND correlation < 0.5;
```

**Solutions**:
1. **Add indexes**: Create indexes on frequently queried columns
2. **Query optimization**: Rewrite complex queries
3. **Database tuning**: Adjust PostgreSQL configuration
4. **Connection pooling**: Reduce connection overhead

### 4. Networking Issues

#### Service Mesh Problems

**Symptoms**:
- Inter-service communication failures
- NetworkPolicy blocking traffic

**Diagnosis**:
```bash
# Check network policies
kubectl get networkpolicies -n arch-separation

# Test service connectivity
kubectl run test-pod --image=busybox --rm -it -- /bin/sh
# Inside pod: wget http://api-gateway:8080/health

# Check DNS resolution
kubectl exec -it test-pod -- nslookup api-gateway.arch-separation.svc.cluster.local
```

**Solutions**:
1. **Update NetworkPolicies**: Ensure proper pod selectors and ports
2. **Check service discovery**: Verify DNS and service definitions
3. **Validate ingress rules**: Confirm ingress configuration
4. **Test from different pods**: Isolate networking issues

#### SSL/TLS Issues

**Symptoms**:
- Certificate errors
- HTTPS redirects failing

**Diagnosis**:
```bash
# Check certificate status
kubectl get certificates -n arch-separation

# Verify cert-manager
kubectl describe certificate api-gateway-tls -n arch-separation

# Test SSL connection
openssl s_client -connect api.your-domain.com:443 -servername api.your-domain.com
```

**Solutions**:
1. **Renew certificates**: Force certificate renewal
2. **Check DNS**: Ensure proper DNS configuration
3. **Validate ingress TLS**: Confirm TLS configuration
4. **Update cert-manager**: Ensure latest version

### 5. Monitoring Issues

#### Missing Metrics

**Symptoms**:
- Grafana dashboards showing no data
- Prometheus targets down

**Diagnosis**:
```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health != "up")'

# Verify service discovery
kubectl get endpoints -n arch-separation

# Check Grafana data sources
# Access Grafana UI and test data source
```

**Solutions**:
1. **Update service annotations**: Add Prometheus scrape annotations
2. **Check network policies**: Allow monitoring traffic
3. **Validate metrics endpoints**: Ensure `/metrics` endpoints respond
4. **Restart monitoring stack**: Redeploy Prometheus/Grafana

#### Alert Fatigue

**Symptoms**:
- Too many alerts
- False positive alerts

**Diagnosis**:
```bash
# Check alert rules
curl http://localhost:9090/api/v1/rules | jq '.data.groups[]'

# Review recent alerts
curl http://localhost:9093/api/v2/alerts | jq '.[] | select(.status.state == "firing")'
```

**Solutions**:
1. **Tune alert thresholds**: Adjust based on normal operation
2. **Implement alert grouping**: Reduce notification noise
3. **Add alert dependencies**: Prevent cascade alerts
4. **Update runbooks**: Improve alert response procedures

## 🛠️ Diagnostic Tools

### Log Analysis

```bash
# Structured log search
kubectl logs deployment/api-gateway -n arch-separation | jq 'select(.level == "error")'

# Time-based log filtering
kubectl logs --since=1h deployment/api-gateway -n arch-separation

# Multi-container log aggregation
stern ".*" -n arch-separation --tail=50
```

### Performance Profiling

```bash
# Go pprof
kubectl port-forward deployment/api-gateway 8080:8080 -n arch-separation
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory analysis
kubectl exec -it deployment/api-gateway -n arch-separation -- /app/binary -memprofile=/tmp/mem.prof
go tool pprof /tmp/mem.prof
```

### Network Debugging

```bash
# Packet capture
kubectl run netshoot --image=nicolaka/netshoot --rm -it
tcpdump -i eth0 host api-gateway

# DNS debugging
kubectl run dnsutils --image=tutum/dnsutils --rm -it
nslookup api-gateway.arch-separation.svc.cluster.local

# Connectivity testing
kubectl run test-pod --image=busybox --rm -it
telnet api-gateway 8080
```

## 📞 Escalation Procedures

### Severity Levels

1. **SEV-1 (Critical)**: Complete service outage, security breach
   - Immediate response required
   - Wake up on-call personnel
   - Customer impact

2. **SEV-2 (High)**: Major functionality broken, performance issues
   - Response within 1 hour
   - Multiple team members involved
   - Significant customer impact

3. **SEV-3 (Medium)**: Minor functionality issues, monitoring alerts
   - Response within 4 hours
   - Single team member can resolve
   - Limited customer impact

4. **SEV-4 (Low)**: Cosmetic issues, informational
   - Response within 24 hours
   - No immediate customer impact

### Escalation Path

1. **First Response**: On-call engineer (15 minutes)
2. **Team Lead**: If unresolved after 1 hour
3. **Engineering Manager**: If unresolved after 4 hours
4. **VP Engineering**: If unresolved after 24 hours

## 📊 Health Check Queries

### Kubernetes Health

```bash
# Pod health summary
kubectl get pods -A --field-selector=status.phase!=Running

# Resource usage
kubectl top nodes
kubectl top pods -n arch-separation

# Event monitoring
kubectl get events -n arch-separation --sort-by='.lastTimestamp' | tail -20
```

### Application Health

```bash
# API health matrix
curl -s https://api.your-domain.com/health | jq .

# Worker health
curl -s https://ai-worker.your-domain.workers.dev/health | jq .

# Database health
kubectl exec -it deployment/postgres -n arch-separation -- pg_isready
```

### Monitoring Health

```bash
# Prometheus health
curl -s http://localhost:9090/-/healthy

# Alertmanager health
curl -s http://localhost:9093/-/healthy

# Grafana health
curl -s http://localhost:3000/api/health
```

## 🔄 Recovery Procedures

### Service Recovery

1. **Identify failed components**: Use monitoring to pinpoint issues
2. **Isolate affected systems**: Prevent cascade failures
3. **Implement fixes**: Apply patches, configuration changes
4. **Gradual rollout**: Test fixes in staging before production
5. **Monitor recovery**: Ensure stability before full restoration

### Data Recovery

1. **Assess data loss**: Determine scope and impact
2. **Restore from backups**: Use latest clean backup
3. **Validate data integrity**: Check restored data consistency
4. **Update affected systems**: Refresh caches and indexes
5. **Monitor system behavior**: Watch for anomalies post-recovery

## 📚 Additional Resources

- [Kubernetes Troubleshooting Guide](https://kubernetes.io/docs/tasks/debug/)
- [Cloudflare Workers Documentation](https://developers.cloudflare.com/workers/)
- [Prometheus Alerting](https://prometheus.io/docs/alerting/latest/)
- [PostgreSQL Performance Tuning](https://www.postgresql.org/docs/current/monitoring.html)

## 📞 Contact Information

- **Emergency Hotline**: +1-XXX-XXX-XXXX
- **Operations Slack**: #platform-ops
- **Email**: ops@your-domain.com
- **On-call Schedule**: https://your-domain.pagerduty.com