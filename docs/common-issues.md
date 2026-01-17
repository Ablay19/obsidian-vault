# Common Issues and Solutions

This document provides solutions to common problems encountered when using Obsidian Vault services.

## WhatsApp Integration

### Issue: Webhook Not Receiving Messages

**Symptoms**
- WhatsApp messages not appearing in logs
- No response to incoming messages
- 403 Forbidden errors on webhook requests

**Solutions**

1. **Verify Webhook URL**
   ```bash
   # Check webhook is correctly configured
   curl -X GET "https://graph.facebook.com/v18.0/YOUR_PHONE_NUMBER_ID?fields= webhook_url&access_token=YOUR_TOKEN"
   ```

2. **Check Verification Token**
   - Ensure `VERIFY_TOKEN` matches exactly in code and Meta dashboard
   - Token is case-sensitive

3. **Verify SSL Certificate**
   - Webhook URL must use HTTPS (except localhost)
   - Certificate must be valid and not self-signed (in production)

4. **Check Meta App Status**
   - Ensure app is in "Live" mode for production
   - Verify WhatsApp product is approved

---

### Issue: Messages Not Being Sent

**Symptoms**
- API returns success but user doesn't receive message
- Messages stuck in "sent" status

**Solutions**

1. **Check Phone Number Status**
   ```bash
   # Verify phone number is approved
   curl -X GET "https://graph.facebook.com/v18.0/YOUR_PHONE_NUMBER_ID?fields=verified_name,quality_rating&access_token=YOUR_TOKEN"
   ```

2. **Review Template Status**
   - Ensure template is approved in Meta dashboard
   - Check template language matches message

3. **Verify Message Type Support**
   - Templates required for initiating conversations
   - Session messages only for 24-hour windows

---

### Issue: Rate Limiting

**Symptoms**
- "Too Many Requests" errors
- Messages delayed or dropped

**Solutions**

1. **Implement Rate Limiting**
   ```go
   // In your middleware
   rateLimiter := NewRateLimiter(60, time.Minute) // 60 messages per minute
   if !rateLimiter.Allow(chatID) {
       return rateLimitExceeded
   }
   ```

2. **Use Message Queues**
   - Implement async message sending
   - Use RabbitMQ or similar for backlog

3. **Monitor Usage**
   - Check Meta Business Manager for API usage
   - Set up alerts for threshold approaching

---

## AI Service Issues

### Issue: AI Responses Slow or Timing Out

**Symptoms**
- Requests take >30 seconds
- Timeout errors in logs

**Solutions**

1. **Check Provider Status**
   ```bash
   # Test OpenAI
   curl https://status.openai.com
   
   # Test Anthropic
   curl https://status.anthropic.com
   ```

2. **Optimize Prompt Length**
   - Reduce unnecessary context
   - Use shorter prompts for simple queries

3. **Implement Caching**
   ```go
   cache := NewLLMCache(24 * time.Hour)
   if cached := cache.Get(prompt); cached != nil {
       return cached
   }
   ```

4. **Use Streaming**
   - Stream responses to reduce perceived latency
   - Set appropriate timeouts

---

### Issue: Poor Response Quality

**Symptoms**
- Inaccurate or irrelevant responses
- Inconsistent formatting

**Solutions**

1. **Improve System Prompt**
   ```json
   {
     "role": "system",
     "content": "You are a specialized assistant. Always structure responses with: 1) Summary 2) Details 3) Examples"
   }
   ```

2. **Add Context**
   - Include relevant documents in prompt
   - Use RAG for domain-specific queries

3. **Adjust Parameters**
   - Increase `temperature` for more creative responses
   - Increase `max_tokens` for longer responses
   - Use better model for complex queries

---

### Issue: Cost Unexpectedly High

**Symptoms**
- Unexpected charges from AI providers
- Token usage spikes

**Solutions**

1. **Implement Token Tracking**
   ```go
   usage := &AIUsage{
       UserID:    userID,
       Model:     model,
       TokensIn:  promptTokens,
       TokensOut: completionTokens,
       Cost:      calculateCost(model, promptTokens, completionTokens),
   }
   db.Save(usage)
   ```

2. **Set Usage Limits**
   ```go
   maxMonthlyCost := 100.0 // USD
   if user.GetMonthlyCost() >= maxMonthlyCost {
       return rateLimited
   }
   ```

3. **Use Cheaper Models**
   - Use GPT-3.5 for simple queries
   - Reserve GPT-4 for complex tasks

---

## Database Issues

### Issue: Connection Pool Exhaustion

**Symptoms**
- "too many connections" errors
- Slow database queries
- Timeouts during high load

**Solutions**

1. **Adjust Pool Size**
   ```go
   db, err := sql.Open("postgres", dsn)
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(5)
   db.SetConnMaxLifetime(5 * time.Minute)
   ```

2. **Implement Query Timeouts**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   db.QueryContext(ctx, "SELECT ...")
   ```

3. **Optimize Long-Running Queries**
   ```sql
   -- Add indexes
   CREATE INDEX idx_user_id ON jobs(user_id);
   
   -- Use EXPLAIN ANALYZE
   EXPLAIN ANALYZE SELECT * FROM jobs WHERE user_id = '123';
   ```

---

### Issue: Data Inconsistency

**Symptoms**
- Missing records
- Duplicate entries
- Stale data

**Solutions**

1. **Implement Transactions**
   ```go
   tx, err := db.Begin()
   _, err = tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = $1", fromID)
   _, err = tx.Exec("UPDATE accounts SET balance = balance + 100 WHERE id = $1", toID)
   if err != nil {
       tx.Rollback()
       return err
   }
   return tx.Commit()
   ```

2. **Add Constraints**
   ```sql
   ALTER TABLE jobs ADD CONSTRAINT fk_session 
   FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE;
   ```

3. **Implement Idempotency**
   ```go
   key := fmt.Sprintf("idempotent:%s", requestID)
   if redis.Exists(ctx, key) {
       return redis.Get(ctx, key)
   }
   result := processRequest()
   redis.Set(ctx, key, result, 24*time.Hour)
   return result
   ```

---

## Cloudflare Worker Issues

### Issue: Worker Not Deploying

**Symptoms**
- `wrangler deploy` fails
- Build errors in Cloudflare dashboard

**Solutions**

1. **Check wrangler.toml**
   ```toml
   name = "ai-manim-worker"
   main = "src/index.ts"
   compatibility_date = "2024-01-01"
   ```

2. **Verify TypeScript Compilation**
   ```bash
   npx tsc --noEmit
   ```

3. **Check Size Limits**
   - Workers have 1MB script size limit
   - Use dynamic imports for large libraries

---

### Issue: KV Operations Failing

**Symptoms**
- `SESSIONS.get()` returns null
- Write operations timeout

**Solutions**

1. **Verify Binding Configuration**
   ```toml
   [[kv_namespaces]]
   binding = "SESSIONS"
   id = "your-kv-namespace-id"
   ```

2. **Check Namespace ID**
   ```bash
   wrangler kv:namespace list
   ```

3. **Implement Retry Logic**
   ```typescript
   async function getWithRetry(key: string, retries = 3): Promise<string | null> {
       for (let i = 0; i < retries; i++) {
           try {
               return await env.SESSIONS.get(key);
           } catch (e) {
               if (i === retries - 1) throw e;
               await new Promise(r => setTimeout(r, 100 * Math.pow(2, i)));
           }
       }
       return null;
   }
   ```

---

## R2 Storage Issues

### Issue: Upload Fails

**Symptoms**
- 403 Forbidden on upload
- Signature validation errors

**Solutions**

1. **Verify Credentials**
   ```bash
   # Test R2 connection
   aws --endpoint-url=https://your-account.r2.cloudflarestorage.com s3 ls
   ```

2. **Check Bucket Permissions**
   - Ensure R2 API token has write permissions
   - Verify CORS configuration

3. **Use Correct Endpoint**
   ```typescript
   const R2_ENDPOINT = `https://${env.CLOUDFLARE_ACCOUNT_ID}.r2.cloudflarestorage.com`;
   ```

---

### Issue: Download URL Expired

**Symptoms**
- Presigned URLs return 403
- "Signature expired" errors

**Solutions**

1. **Increase URL Validity**
   ```typescript
   const url = await env.VIDEO_BUCKET.createSignedUrl(
       'get_object',
       key,
       15 * 60 // 15 minutes
   );
   ```

2. **Synchronize Time**
   - Ensure server time is accurate
   - Use NTP synchronization

3. **Implement URL Refresh**
   ```typescript
   async function getAccessibleUrl(key: string): Promise<string> {
       const url = await createSignedUrl(key);
       if (isExpiringSoon(url)) {
           return await refreshUrl(key);
       }
       return url;
   }
   ```

---

## General Issues

### Issue: High Memory Usage

**Symptoms**
- Out of memory errors
- Slow response times
- Worker restarts

**Solutions**

1. **Profile Memory Usage**
   ```go
   import "runtime/pprof"
   
   f, _ := os.Create("mem.prof")
   pprof.WriteHeapProfile(f)
   f.Close()
   ```

2. **Optimize Data Structures**
   - Use streaming for large data
   - Avoid keeping references

3. **Implement Limits**
   ```go
   const MaxRequestSize = 10 * 1024 * 1024 // 10MB
   if size > MaxRequestSize {
       return RequestTooLarge
   }
   ```

---

### Issue: Logging Not Working

**Symptoms**
- No logs in Cloudflare dashboard
- Missing error details

**Solutions**

1. **Check Log Level**
   ```typescript
   const logger = createLogger({
       level: env.LOG_LEVEL || 'info'
   });
   ```

2. **Use Correct Destination**
   - Workers logs appear in Cloudflare dashboard
   - Use `console.log()` for basic output

3. **Implement Structured Logging**
   ```go
   logger.Info("job completed",
       zap.String("job_id", jobID),
       zap.Duration("duration", duration),
       zap.Error(err),
   )
   ```

---

## Getting Help

If your issue is not listed here:

1. **Check Logs**
   - Review application logs for detailed error messages
   - Check Cloudflare Worker logs
   - Monitor system metrics

2. **Search Documentation**
   - Review relevant service documentation
   - Check API reference for endpoint requirements

3. **Contact Support**
   - Internal Slack: #engineering-support
   - Email: support@obsidianvault.com
   - Include error logs and reproduction steps
