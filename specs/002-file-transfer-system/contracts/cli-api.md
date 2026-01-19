# Secure File Transfer System CLI API

## Overview

The file transfer system extends the Mauritania CLI with new commands for secure binary file transfers across existing transports. Commands follow the established CLI patterns with progress reporting and error handling.

## Commands

### Upload Command

**Command**: `mauritania-cli upload <file-path> [options]`

**Description**: Upload a file to the remote environment using available transports

**Parameters**:
- `file-path`: Local file path to upload (required)
- `--transport <type>`: Preferred transport (whatsapp, telegram, facebook, smapos)
- `--compression <algo>`: Compression algorithm (gzip, zstd, none)
- `--encrypt`: Enable end-to-end encryption
- `--chunk-size <size>`: Chunk size in MB (default: 1)
- `--priority <level>`: Transfer priority 1-10 (default: 5)

**Options**:
- `--dry-run`: Show what would be uploaded without transferring
- `--progress`: Show detailed progress (default: enabled)
- `--resume`: Attempt to resume interrupted transfer
- `--timeout <seconds>`: Transfer timeout (default: 3600)

**Output**:
```
Starting upload of large-model.bin (2.5 GB)
Using transport: whatsapp
Compression: gzip
Chunk size: 1 MB
Total chunks: 2500

Progress: [████████████████████████] 100% | 2.5 GB / 2.5 GB
Transfer completed successfully
Transfer ID: abc-123-def
Remote path: /uploads/large-model.bin.gz
```

**Error Codes**:
- 1: File not found
- 2: Insufficient permissions
- 3: Transport unavailable
- 4: Transfer timeout
- 5: Integrity verification failed

### Download Command

**Command**: `mauritania-cli download <remote-path> <local-path> [options]`

**Description**: Download a file from the remote environment

**Parameters**:
- `remote-path`: Remote file path to download (required)
- `local-path`: Local destination path (required)
- `--transport <type>`: Preferred transport
- `--decompress`: Automatically decompress if compressed
- `--decrypt`: Decrypt if encrypted

**Options**:
- `--resume`: Resume interrupted download
- `--overwrite`: Overwrite existing local file
- `--progress`: Show progress (default: enabled)

**Output**:
```
Starting download of /logs/app.log
From transport: telegram
File size: 50 MB

Progress: [██████████████░░░░░░░░] 60% | 30 MB / 50 MB
Download resumed from chunk 1500/2500

Progress: [████████████████████████] 100% | 50 MB / 50 MB
Download completed successfully
Local path: ./app.log
SHA-256: a1b2c3d4...
```

### Transfer Status Command

**Command**: `mauritania-cli transfer status [transfer-id]`

**Description**: Check status of active or completed transfers

**Parameters**:
- `transfer-id`: Specific transfer ID (optional, shows all if omitted)

**Options**:
- `--active`: Show only active transfers
- `--completed`: Show only completed transfers
- `--failed`: Show only failed transfers

**Output**:
```
Active Transfers:
┌─────────────────────────────────────┬────────────┬──────────┬────────────┐
│ Transfer ID                        │ File       │ Progress │ Status     │
├─────────────────────────────────────┼────────────┼──────────┼────────────┤
│ abc-123-def-456                    │ model.bin  │ 45%      │ Transferring│
│ ghi-789-jkl-012                    │ data.csv   │ 100%     │ Reassembling│
└─────────────────────────────────────┴────────────┴──────────┴────────────┘

Completed Transfers (Last 24h):
- model.bin (2.5 GB) completed 2 hours ago
- config.yaml (1.2 KB) completed 6 hours ago
```

### Transfer Cancel Command

**Command**: `mauritania-cli transfer cancel <transfer-id>`

**Description**: Cancel an active transfer

**Parameters**:
- `transfer-id`: Transfer ID to cancel (required)

**Output**:
```
Transfer abc-123-def-456 cancelled successfully
Cleaned up 1250 transferred chunks
```

### Transfer List Command

**Command**: `mauritania-cli transfer list [filters]`

**Description**: List all transfers with filtering options

**Options**:
- `--user <username>`: Filter by user
- `--status <status>`: Filter by status (pending, active, completed, failed)
- `--transport <type>`: Filter by transport
- `--since <date>`: Show transfers since date
- `--limit <number>`: Limit results (default: 50)

**Output**: Tabular list similar to status command

## Transport Interface Extensions

### Binary Send Method

```go
type Transport interface {
    // Existing methods...
    SendBinary(ctx context.Context, data []byte, metadata map[string]interface{}) error
    ReceiveBinary(ctx context.Context, sessionID string) ([]byte, error)
    CreateSession(ctx context.Context, config SessionConfig) (string, error)
    CloseSession(ctx context.Context, sessionID string) error
}
```

### Session Configuration

```go
type SessionConfig struct {
    TransferID   string
    Direction    TransferDirection
    FileSize     int64
    ChunkCount   int
    Compression  string
    Encryption   bool
    Priority     int
    Timeout      time.Duration
}
```

## Progress Reporting

### Progress Callback Interface

```go
type ProgressCallback func(progress TransferProgress)

type TransferProgress struct {
    TransferID       string
    TotalChunks      int
    CompletedChunks  int
    CurrentChunk     int
    BytesTransferred int64
    TotalBytes       int64
    Status           TransferStatus
    EstimatedTimeRemaining time.Duration
    CurrentSpeed     float64 // bytes per second
}
```

### CLI Progress Display

- Real-time progress bar with percentage
- Transfer speed and ETA
- Chunk completion status
- Error notifications with retry information
- Completion summary with integrity verification

## Error Handling

### Transport-Specific Errors

- **Message Too Large**: Automatic chunk size reduction
- **Rate Limited**: Exponential backoff retry
- **Connection Lost**: Automatic resume capability
- **Authentication Failed**: Clear error with resolution steps

### Recovery Mechanisms

- **Automatic Retry**: Failed chunks resent with backoff
- **Manual Resume**: User can resume interrupted transfers
- **Partial Cleanup**: Failed transfers cleaned up safely
- **Integrity Checks**: Corrupted data detected and re-sent

## Configuration

### CLI Configuration File

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

### Environment Variables

```bash
MAURITANIA_TRANSFER_ENCRYPTION_KEY=your-encryption-key
MAURITANIA_TRANSFER_STORAGE_PATH=/tmp/transfers
MAURITANIA_TRANSFER_MAX_FILE_SIZE=1073741824  # 1GB
```

## Security Considerations

- File type validation prevents malicious uploads
- Transport encryption protects data in transit
- Optional end-to-end encryption for sensitive files
- Access logging for audit trails
- Automatic cleanup of temporary files

## Performance Characteristics

- **Chunk Size**: 1MB optimal for most transports
- **Concurrency**: Up to 3 simultaneous transfers
- **Resume Capability**: Sub-30 second recovery time
- **Compression Ratio**: 60-80% size reduction for text/binary files
- **Integrity Verification**: <5 seconds for 100MB files