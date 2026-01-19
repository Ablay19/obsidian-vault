# Secure File Transfer System Quickstart

## Overview

Get started with secure binary file transfers over unreliable network connections using the Mauritania CLI. This feature enables uploading and downloading files through messaging platforms and specialized shippers.

## Prerequisites

- Mauritania CLI installed and configured
- At least one transport configured (WhatsApp, Telegram, Facebook, or SM APOS Shipper)
- Files to transfer (up to 100MB per file)

## Basic Usage

### Upload a File

```bash
# Upload a binary file
mauritania-cli upload my-model.bin

# Upload with specific transport
mauritania-cli upload data.csv --transport whatsapp

# Upload with compression
mauritania-cli upload large-file.zip --compression gzip

# Upload with custom chunk size
mauritania-cli upload big-file.dat --chunk-size 2
```

### Download a File

```bash
# Download from remote path
mauritania-cli download /remote/path/model.bin ./local/model.bin

# Download with automatic decompression
mauritania-cli download /logs/app.log.gz ./logs/app.log --decompress

# Resume interrupted download
mauritania-cli download /data/dataset.csv ./data/dataset.csv --resume
```

### Monitor Transfers

```bash
# Check active transfers
mauritania-cli transfer status

# Check specific transfer
mauritania-cli transfer status abc-123-def

# List all transfers
mauritania-cli transfer list

# List failed transfers
mauritania-cli transfer list --failed
```

## Advanced Usage

### Encrypted Transfers

```bash
# Upload with end-to-end encryption
mauritania-cli upload sensitive-data.zip --encrypt

# Download encrypted file
mauritania-cli download /secure/data.enc ./data.zip --decrypt
```

### Priority Transfers

```bash
# High priority upload
mauritania-cli upload critical-update.bin --priority 9

# Low priority background transfer
mauritania-cli upload backup.tar --priority 1
```

### Batch Operations

```bash
# Upload multiple files (bash)
for file in *.bin; do
  mauritania-cli upload "$file" --transport whatsapp &
done

# Download multiple files
mauritania-cli download /batch/file1.bin ./downloads/
mauritania-cli download /batch/file2.bin ./downloads/
```

## Configuration

### Environment Variables

```bash
# Encryption key for sensitive transfers
export MAURITANIA_TRANSFER_ENCRYPTION_KEY="your-32-char-key"

# Temporary storage path
export MAURITANIA_TRANSFER_STORAGE_PATH="/tmp/mauritania-transfers"

# Maximum file size
export MAURITANIA_TRANSFER_MAX_FILE_SIZE="1073741824"  # 1GB

# Default transport
export MAURITANIA_TRANSFER_DEFAULT_TRANSPORT="whatsapp"
```

### CLI Configuration

Create `~/.mauritania/transfer-config.yaml`:

```yaml
file-transfer:
  default-transport: whatsapp
  default-chunk-size: 1048576  # 1MB
  default-compression: gzip
  max-concurrent-transfers: 3
  transfer-timeout: 3600
  retry-attempts: 3
  retry-delay: 30
  enable-encryption: false
  progress-update-interval: 5
```

## Transport-Specific Setup

### WhatsApp Transport

```bash
# Ensure WhatsApp is connected
mauritania-cli status

# Test file transfer
mauritania-cli upload test.txt --transport whatsapp
```

### Telegram Transport

```bash
# Configure Telegram bot token
export TELEGRAM_BOT_TOKEN="your-bot-token"

# Upload file
mauritania-cli upload document.pdf --transport telegram
```

### SM APOS Shipper

```bash
# Configure shipper credentials
export SMAPOS_API_KEY="your-api-key"

# Upload large file
mauritania-cli upload large-dataset.zip --transport smapos
```

## Troubleshooting

### Common Issues

**Transfer Stuck at 0%**
```bash
# Check transport connectivity
mauritania-cli status

# Verify transport configuration
mauritania-cli config list
```

**File Corruption**
```bash
# Check integrity logs
mauritania-cli transfer status <transfer-id>

# Re-download if corrupted
mauritania-cli download <remote-path> <local-path> --overwrite
```

**Slow Transfers**
```bash
# Reduce chunk size
mauritania-cli upload large-file.bin --chunk-size 0.5

# Try different transport
mauritania-cli upload large-file.bin --transport smapos
```

**Permission Denied**
```bash
# Check file permissions
ls -la <file-path>

# Verify user has transport access
mauritania-cli whoami
```

### Recovery from Interruptions

```bash
# Resume interrupted upload
mauritania-cli upload large-file.bin --resume

# Resume interrupted download
mauritania-cli download /remote/file.bin ./local/file.bin --resume

# Check resumable transfers
mauritania-cli transfer list --active
```

## Performance Tips

### Optimize for Speed

- Use SM APOS Shipper for large files (>10MB)
- Enable compression for text/binary files
- Increase chunk size for reliable connections
- Avoid encryption unless required

### Optimize for Reliability

- Use smaller chunks for unstable connections
- Enable retry with backoff
- Monitor transfer status regularly
- Keep multiple transports configured

### Storage Management

```bash
# Clean up completed transfers
find /tmp/mauritania-transfers -name "*.chunk" -mtime +1 -delete

# Monitor storage usage
du -sh /tmp/mauritania-transfers

# Configure retention policy
export MAURITANIA_TRANSFER_RETENTION_DAYS=7
```

## Security Best Practices

- Enable encryption for sensitive files
- Use strong encryption keys
- Regularly rotate transport credentials
- Monitor transfer logs for anomalies
- Clean up temporary files after transfers

## Examples

### ML Model Deployment

```bash
# Train model locally
python train.py

# Compress model
gzip model.pkl

# Upload to remote
mauritania-cli upload model.pkl.gz --transport whatsapp --encrypt

# Deploy on remote
mauritania-cli send "python deploy.py"
```

### Log Analysis

```bash
# Download application logs
mauritania-cli download /var/log/app.log ./analysis/

# Download multiple log files
for log in access error app; do
  mauritania-cli download "/var/log/${log}.log" "./analysis/"
done

# Analyze logs locally
python analyze_logs.py ./analysis/*.log
```

### Configuration Backup

```bash
# Backup remote configurations
mauritania-cli download /etc/nginx/nginx.conf ./backup/
mauritania-cli download /etc/app/config.yaml ./backup/

# Restore configurations
mauritania-cli upload ./backup/nginx.conf /etc/nginx/
mauritania-cli upload ./backup/config.yaml /etc/app/
```

## Integration with CI/CD

### GitHub Actions

```yaml
- name: Deploy Binary
  run: |
    mauritania-cli upload app-binary --transport whatsapp
    mauritania-cli send "systemctl restart app"
```

### Jenkins Pipeline

```groovy
stage('Deploy') {
    steps {
        sh 'mauritania-cli upload dist/app.jar --transport telegram'
        sh 'mauritania-cli send "java -jar /opt/app/app.jar"'
    }
}
```

For detailed documentation, visit the [File Transfer Guide](https://docs.mauritania-cli.dev/file-transfer).