# WhatsApp Business API Integration Guide

## 📋 Table of Contents
1. [Prerequisites](#prerequisites)
2. [Meta Developer Account Setup](#meta-developer-account-setup)
3. [WhatsApp Business App Configuration](#whatsapp-business-app-configuration)
4. [Environment Variable Setup](#environment-variable-setup)
5. [Webhook Configuration](#webhook-configuration)
6. [Testing & Validation](#testing--validation)
7. [Dashboard Features](#dashboard-features)
8. [Troubleshooting](#troubleshooting)
9. [Security Best Practices](#security-best-practices)
10. [Production Deployment](#production-deployment)

---

## 🚀 Prerequisites

### Technical Requirements
- **HTTPS Domain**: All webhook URLs must use HTTPS with valid SSL certificate
- **Server Access**: Publicly accessible server for webhook endpoints
- **File Storage**: Local filesystem for media attachments (default: `attachments/whatsapp/`)
- **Database**: SQLite or Turso database for configuration storage
- **Go Dependencies**: All WhatsApp-related dependencies are included in the project

### Business Requirements
- **Meta Business Account**: Verified Meta for Developers account
- **WhatsApp Business Account**: Verified WhatsApp Business API account
- **Phone Number**: Dedicated phone number for WhatsApp Business
- **Business Verification**: Verified business information on Meta

### Software Dependencies
```bash
# Required Go packages (already included)
github.com/go-telegram-bot-api/telegram-bot-api/v5
go.uber.org/zap
github.com/uptrace/opentelemetry-go-extra/otelzap
modernc.org/sqlite
github.com/tursodatabase/libsql-client-go/libsql

# External binaries for media processing
tesseract-ocr      # For text extraction from images
pdftotext         # For text extraction from PDFs
```

---

## 🔧 Meta Developer Account Setup

### Step 1: Create Meta Developer Account
1. **Visit**: [developers.facebook.com](https://developers.facebook.com)
2. **Sign Up**: Create account using business email
3. **Verify Identity**: Complete business verification process
4. **Set Up 2FA**: Enable two-factor authentication

### Step 2: Create Meta App
1. **Navigate**: My Apps → Create App
2. **Select App Type**: Choose "Business"
3. **App Details**:
   ```
   App Name: Obsidian Automation Bot
   App Contact: your-business-email@domain.com
   Business Portfolio: Select your business
   ```
4. **Create App**: Click "Create App"

### Step 3: Add WhatsApp Product
1. **Find WhatsApp**: In app dashboard → Products → Find "WhatsApp"
2. **Click Add**: Select "WhatsApp" from product catalog
3. **Choose Setup Method**: Select "Business"
4. **Business Verification**: Link your WhatsApp Business account

---

## 📱 WhatsApp Business App Configuration

### Step 1: Phone Number Setup
1. **Link Phone Number**: Connect your dedicated WhatsApp number
2. **Verify Number**: Complete verification process (SMS/call)
3. **Display Name**: Set your business display name (approved by WhatsApp)
4. **Business Profile**: Complete business information

### Step 2: API Configuration
1. **Navigate**: WhatsApp → API Setup → Configuration
2. **Webhook URL**: Set your webhook endpoint
   ```
   Production: https://your-domain.com/api/v1/auth/whatsapp/webhook
   Testing:    https://your-domain.com/api/v1/auth/whatsapp/webhook
   ```
3. **Verify Token**: Set your custom verification token

### Step 3: Get Required Credentials

#### Access Token
1. **Location**: WhatsApp → API Setup → Configuration
2. **Generate**: Click "Generate" or "Reset" for permanent token
3. **Copy**: Safely copy the generated token
4. **Store**: Save to your environment variables immediately

#### Verify Token
1. **Location**: WhatsApp → API Setup → Webhook
2. **Create**: Generate your custom verify token
3. **Security**: Use a strong, unique token

#### App Secret
1. **Location**: App Settings → Basic → App Secret
2. **Reveal**: Click "Show" to display the secret
3. **Copy**: Carefully copy the secret value
4. **Secure**: Store in secure environment variables

---

## 🔑 Environment Variable Setup

### Configuration File Setup
Create or update your `.env` file with WhatsApp credentials:

```bash
# WhatsApp Business API Configuration
WHATSAPP_ACCESS_TOKEN="EAADZ...your-long-access-token"
WHATSAPP_VERIFY_TOKEN="your-custom-verify-token-2024"
WHATSAPP_APP_SECRET="your-app-secret-from-meta-dashboard"
WHATSAPP_WEBHOOK_URL="https://your-domain.com/api/v1/auth/whatsapp/webhook"
```

### Environment Variable Descriptions

| Variable | Required | Description | Where to Find |
|----------|-----------|-------------|----------------|
| `WHATSAPP_ACCESS_TOKEN` | ✅ Yes | Permanent API token for WhatsApp Business API | WhatsApp → API Setup → Configuration |
| `WHATSAPP_VERIFY_TOKEN` | ✅ Yes | Custom token for webhook verification challenge | WhatsApp → API Setup → Webhook |
| `WHATSAPP_APP_SECRET` | ✅ Yes | App secret for webhook signature validation | App Settings → Basic → App Secret |
| `WHATSAPP_WEBHOOK_URL` | ✅ Yes | Full URL to your webhook endpoint | Your deployment configuration |

### Loading Configuration
The system automatically loads environment variables using godotenv:

```go
// In cmd/bot/main.go
config.LoadConfig() // Loads .env file into OS environment
```

---

## 🔗 Webhook Configuration

### Webhook Endpoints

#### Incoming Messages (POST)
```
POST /api/v1/auth/whatsapp/webhook
Content-Type: application/json
X-Hub-Signature-256: sha256=...hmac-signature
```

#### Webhook Verification (GET)
```
GET /api/v1/auth/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=YOUR_TOKEN&hub.challenge=RANDOM_STRING
```

### Webhook Security Implementation

#### HMAC-SHA256 Signature Validation
The system implements secure webhook validation:

```go
func (h *Handler) validateSignature(signature string, payload []byte) bool {
    // Extract signature from header
    actualSignature := strings.TrimPrefix(signature, "sha256=")
    
    // Get app secret from configuration
    secret := h.service.GetConfig().AppSecret
    
    // Calculate expected signature
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    
    // Compare signatures securely
    return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}
```

### Webhook Event Types

#### Message Events
```json
{
  "object": "whatsapp_business_account",
  "entry": [{
    "id": "WEBHOOK_ID",
    "changes": [{
      "field": "messages",
      "value": {
        "messaging_product": "whatsapp",
        "metadata": {
          "display_phone_number": "+1234567890",
          "phone_number_id": "1234567890"
        },
        "messages": [{
          "id": "MESSAGE_ID",
          "from": "SENDER_PHONE",
          "timestamp": "TIMESTAMP",
          "type": "text|image|audio|video|document",
          "content": {
            "text": {"body": "Message content"}  // For text messages
            // OR
            "image": {  // For images
              "id": "MEDIA_ID",
              "mime_type": "image/jpeg",
              "file_size": 1024000
            }
          }
        }]
      }
    }]
  }]
}
```

---

## 🧪 Testing & Validation

### Step 1: Webhook Verification Test
```bash
# Test webhook verification
curl -X GET "https://your-domain.com/api/v1/auth/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=YOUR_VERIFY_TOKEN&hub.challenge=test123" \
  -v

# Expected response: HTTP 200 with "test123" as body
```

### Step 2: Webhook Security Test
```bash
# Test signature validation with sample payload
curl -X POST "https://your-domain.com/api/v1/auth/whatsapp/webhook" \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=CALCULATED_HMAC" \
  -d '{"object":"whatsapp_business_account","entry":[]}' \
  -v
```

### Step 3: Message Reception Test
1. **Send Test Message**: Send a message to your WhatsApp Business number
2. **Check Logs**: Monitor application logs for incoming messages:
   ```json
   {
     "level": "info",
     "msg": "Handling message",
     "chat_id": "SENDER_PHONE",
     "text": "Received message content"
   }
   ```

### Step 4: Media Download Test
1. **Send Media**: Send an image or document to your WhatsApp number
2. **Verify Storage**: Check local storage directory:
   ```
   attachments/whatsapp/2024/01/06/
   ├── MEDIA_ID_hash16.jpg
   └── MEDIA_ID_hash16.pdf
   ```

---

## 📊 Dashboard Features

### WhatsApp Management Panel
Access at: `http://localhost:8080/dashboard/whatsapp`

#### Connection Status
- **Live Status**: Shows real-time webhook connection state
- **QR Code Display**: For session linking (if applicable)
- **Reconnection Button**: Manual reconnection functionality

#### Monitoring Features
- **WebSocket Integration**: Real-time updates without page refresh
- **Message Logging**: Visual log of incoming messages
- **Error Tracking**: Display of webhook processing errors

### API Endpoints

#### Provider Status
```bash
GET /api/ai/providers
Response:
{
  "available": ["Cloudflare", "Gemini"],
  "active": "Cloudflare"
}
```

#### Service Status
```bash
GET /api/services/status
Response:
{
  "whatsapp": {
    "connected": true,
    "webhook_url": "https://your-domain.com/api/v1/auth/whatsapp/webhook",
    "last_message": "2024-01-06T16:00:00Z"
  }
}
```

---

## 🚨 Troubleshooting

### Common Setup Issues

#### Webhook Verification Fails
**Symptoms**: `403 Forbidden` response during webhook setup
**Causes**: 
- Invalid verify token
- Mismatched webhook URL
- HTTP instead of HTTPS

**Solutions**:
```bash
# Check environment variables
echo "Verify Token: $WHATSAPP_VERIFY_TOKEN"
echo "Webhook URL: $WHATSAPP_WEBHOOK_URL"

# Test manually with curl
curl -X GET "$WHATSAPP_WEBHOOK_URL?hub.mode=subscribe&hub.verify_token=$WHATSAPP_VERIFY_TOKEN&hub.challenge=test123"
```

#### Signature Validation Fails
**Symptoms**: Messages rejected with signature errors
**Causes**:
- Incorrect app secret
- Header parsing issues
- Request body modification

**Debugging**:
```bash
# Check app secret
echo "App Secret: $WHATSAPP_APP_SECRET"

# Test signature calculation
echo -n '{"test":"payload"}' | openssl dgst -sha256 -hmac "$WHATSAPP_APP_SECRET" -binary | openssl enc -base64
```

#### Media Download Fails
**Symptoms**: `404 Not Found` or `403 Forbidden` for media URLs
**Causes**:
- Expired media ID
- Invalid access token
- File size exceeding limits

**Solutions**:
- Verify access token is valid and not expired
- Check media file size (max 50MB default)
- Ensure proper MIME type handling

#### Connection Issues
**Symptoms**: No incoming messages or timeout errors
**Causes**:
- Firewall blocking webhook URL
- SSL certificate issues
- Network connectivity problems

**Debugging**:
```bash
# Test webhook accessibility
curl -X POST "$WHATSAPP_WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{"test": "connection"}' \
  -v

# Check SSL certificate
openssl s_client -connect your-domain.com:443 -servername your-domain.com
```

### Debug Mode Configuration
Enable detailed WhatsApp logging:
```bash
# Add to .env
LOG_LEVEL="debug"
WHATSAPP_DEBUG="true"
```

---

## 🔒 Security Best Practices

### Credential Management
1. **Environment Variables Only**: Never commit credentials to version control
2. **Secure Storage**: Use secret management systems in production
3. **Regular Rotation**: Rotate access tokens every 90 days
4. **Least Privilege**: Use app-scoped permissions only

### Webhook Security
1. **HTTPS Only**: Always use SSL certificates for webhooks
2. **Signature Validation**: Never disable HMAC signature verification
3. **Request Limiting**: Implement rate limiting per IP
4. **Request Logging**: Log all webhook requests for audit trails

### Data Protection
1. **Media Sanitization**: Scan uploaded files for malware
2. **Size Limits**: Enforce reasonable file size restrictions
3. **Type Validation**: Only allow whitelisted MIME types
4. **Secure Storage**: Store media files outside web root

### Network Security
1. **Firewall Rules**: Restrict webhook URL to Meta's IP ranges
2. **VPN Considerations**: Account for legitimate VPN usage
3. **DDoS Protection**: Implement request rate limiting
4. **SSL Configuration**: Use strong TLS configurations

---

## 🚀 Production Deployment

### SSL Certificate Setup
```bash
# Using Let's Encrypt (recommended)
sudo certbot --nginx -d your-domain.com

# Manual certificate installation
sudo cp fullchain.pem /etc/ssl/certs/your-domain.pem
sudo cp privkey.pem /etc/ssl/private/your-domain.key
```

### Nginx Configuration
```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /etc/ssl/certs/your-domain.pem;
    ssl_certificate_key /etc/ssl/private/your-domain.key;
    
    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    # Webhook endpoint
    location /api/v1/auth/whatsapp/webhook {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        
        # Size limits for media
        client_max_body_size 50M;
    }
    
    # Dashboard access
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/bot .

# Create media storage directory
RUN mkdir -p /root/attachments/whatsapp

EXPOSE 8080
CMD ["./bot"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  obsidian-bot:
    build: .
    ports:
      - "8080:8080"
    environment:
      - WHATSAPP_ACCESS_TOKEN=${WHATSAPP_ACCESS_TOKEN}
      - WHATSAPP_VERIFY_TOKEN=${WHATSAPP_VERIFY_TOKEN}
      - WHATSAPP_APP_SECRET=${WHATSAPP_APP_SECRET}
      - WHATSAPP_WEBHOOK_URL=https://your-domain.com/api/v1/auth/whatsapp/webhook
    volumes:
      - ./attachments:/root/attachments
    restart: unless-stopped
```

### Monitoring & Alerting
```bash
# Health check script
#!/bin/bash
WEBHOOK_URL="https://your-domain.com/api/v1/services/status"
RESPONSE=$(curl -s "$WEBHOOK_URL")

if echo "$RESPONSE" | grep -q '"whatsapp".*"connected":true'; then
    echo "✅ WhatsApp service healthy"
else
    echo "❌ WhatsApp service down"
    # Send alert
    curl -X POST "https://hooks.slack.com/YOUR_WEBHOOK" \
      -d '{"text":"🚨 WhatsApp service is down!"}'
fi
```

---

## 📚 API Reference

### Message Types

```go
const (
    MessageTypeText     MessageType = "text"
    MessageTypeImage    MessageType = "image"
    MessageTypeAudio    MessageType = "audio"
    MessageTypeVideo    MessageType = "video"
    MessageTypeDocument MessageType = "document"
)
```

### Configuration Structures

```go
// WhatsApp Config
type Config struct {
    AccessToken string
    VerifyToken string
    AppSecret   string
    WebhookURL  string
}

// Message Structure
type Message struct {
    ID        string                 `json:"id"`
    From      string                 `json:"from"`
    Type      MessageType            `json:"type"`
    Timestamp int64                  `json:"timestamp"`
    Content   map[string]interface{} `json:"content"`
}
```

### Error Codes

| Error Code | Description | Solution |
|-------------|-------------|-----------|
| 401 Unauthorized | Invalid access token | Verify WHATSAPP_ACCESS_TOKEN |
| 403 Forbidden | Webhook verification failed | Check WHATSAPP_VERIFY_TOKEN |
| 404 Not Found | Media not found | Check media ID validity |
| 413 Too Large | File size exceeds limit | Reduce file size or increase limits |
| 429 Too Many Requests | Rate limit exceeded | Implement backoff retry |

---

## 🎯 Quick Start Checklist

### Pre-Deployment Checklist
- [ ] Meta Developer account created and verified
- [ ] WhatsApp Business account linked
- [ ] Phone number verified and added
- [ ] Business display name approved
- [ ] All credentials generated and copied
- [ ] Environment variables configured
- [ ] HTTPS certificate installed
- [ ] Firewall rules configured
- [ ] Media storage directory created
- [ ] Database initialized

### Post-Deployment Verification
- [ ] Webhook URL accessible via HTTPS
- [ ] Webhook verification passes test
- [ ] Message reception working
- [ ] Media files downloading correctly
- [ ] Dashboard panel accessible
- [ ] Error logging functional
- [ ] SSL certificate valid
- [ ] Monitoring active

---

## 🔗 Useful Resources

### Official Documentation
- [Meta for Developers](https://developers.facebook.com/docs/whatsapp/)
- [WhatsApp Business API](https://developers.facebook.com/docs/whatsapp/business-api/)
- [Webhook Security](https://developers.facebook.com/docs/graph-api/webhooks/)

### Tools & Utilities
- [Webhook Testing Tools](https://webhook.site/)
- [HMAC Calculator](https://www.freeformatter.com/hmac-generator.html)
- [SSL Test](https://www.ssllabs.com/ssltest/)
- [CURL Generator](https://curlconverter.com/)

### Community Support
- [Meta Developer Community](https://developers.facebook.com/community/)
- [WhatsApp Business Help Center](https://faq.whatsapp.com/)
- [Stack Overflow](https://stackoverflow.com/questions/tagged/whatsapp)

---

**⚡ This comprehensive guide covers everything from initial setup to production deployment. Follow each step carefully to ensure secure and reliable WhatsApp Business API integration.**