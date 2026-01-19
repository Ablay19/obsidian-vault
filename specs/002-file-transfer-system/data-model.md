# Secure File Transfer System Data Model

## Overview

The data model supports secure chunked file transfers across unreliable transport mechanisms with integrity verification, resumable downloads, and progress tracking. All entities include comprehensive validation and support for concurrent transfers.

## Entities

### FileTransfer Entity

**Purpose**: Core transfer metadata and state management for file uploads/downloads

**Fields**:
- `id`: UUID (primary key)
- `user_id`: UUID (foreign key to User, required)
- `direction`: Enum (upload, download, required)
- `local_path`: String (nullable, local file path for uploads)
- `remote_path`: String (nullable, remote file path for downloads)
- `filename`: String (original filename, required, 1-255 chars)
- `file_size`: BigInteger (total file size in bytes, required)
- `file_type`: String (MIME type, required)
- `total_chunks`: Integer (total number of chunks, required)
- `completed_chunks`: Integer (chunks successfully transferred, default 0)
- `file_checksum`: String (SHA-256 of complete file, required)
- `transport_type`: String (transport used: whatsapp, telegram, etc., required)
- `session_id`: String (transport session identifier, nullable)
- `status`: Enum (pending, chunking, transferring, reassembling, completed, failed, cancelled, required)
- `priority`: Integer (transfer priority 1-10, default 5)
- `compression`: String (compression algorithm used, nullable)
- `encryption`: Boolean (end-to-end encryption enabled, default false)
- `created_at`: DateTime (auto-generated)
- `updated_at`: DateTime (auto-updated)
- `started_at`: DateTime (nullable)
- `completed_at`: DateTime (nullable)
- `error_message`: String (nullable, last error description)

**Validation Rules**:
- File size must be positive and within user limits
- Total chunks must equal completed_chunks when status=completed
- File checksum must be valid SHA-256
- Status transitions must be valid (pending → chunking → transferring → reassembling → completed)
- Priority must be 1-10

**Relationships**:
- Many-to-one with User (owning user)
- One-to-many with FileChunk (chunks of this transfer)
- One-to-one with TransferSession (current session)

### FileChunk Entity

**Purpose**: Individual file chunks with integrity verification and transfer state

**Fields**:
- `id`: UUID (primary key)
- `transfer_id`: UUID (foreign key to FileTransfer, required)
- `chunk_index`: Integer (sequential chunk number, required, 0-based)
- `chunk_size`: Integer (size of this chunk in bytes, required)
- `chunk_checksum`: String (SHA-256 of chunk data, required)
- `data`: Blob (compressed chunk data, required)
- `status`: Enum (pending, sent, acknowledged, failed, required)
- `attempts`: Integer (number of send attempts, default 0)
- `sent_at`: DateTime (nullable)
- `acknowledged_at`: DateTime (nullable)
- `error_message`: String (nullable, last transfer error)

**Validation Rules**:
- Chunk index must be unique per transfer and sequential
- Chunk size must be positive and ≤ configured max chunk size
- Chunk checksum must be valid SHA-256
- Status transitions: pending → sent → acknowledged/failed
- Attempts must be non-negative, max retry limit

**Relationships**:
- Many-to-one with FileTransfer (parent transfer)
- Indexes on (transfer_id, chunk_index) for efficient reassembly

### TransferSession Entity

**Purpose**: Transport session management for resumable transfers

**Fields**:
- `id`: UUID (primary key)
- `transfer_id`: UUID (foreign key to FileTransfer, required)
- `transport_type`: String (transport mechanism, required)
- `session_token`: String (transport session identifier, required)
- `session_data`: JSON (transport-specific session state, nullable)
- `status`: Enum (active, paused, expired, terminated, required)
- `created_at`: DateTime (auto-generated)
- `expires_at`: DateTime (required)
- `last_activity`: DateTime (auto-updated)

**Validation Rules**:
- Session token must be unique per transport type
- Expires at must be in future for active sessions
- Status transitions must be valid (active ↔ paused, any → expired/terminated)

**Relationships**:
- One-to-one with FileTransfer (associated transfer)
- Indexes on session_token for quick lookups

### TransferLog Entity

**Purpose**: Audit trail for transfer operations and troubleshooting

**Fields**:
- `id`: UUID (primary key)
- `transfer_id`: UUID (foreign key to FileTransfer, required)
- `chunk_id`: UUID (foreign key to FileChunk, nullable)
- `event_type`: Enum (transfer_started, chunk_sent, chunk_acknowledged, transfer_completed, error_occurred, required)
- `message`: String (event description, required)
- `metadata`: JSON (additional event data, nullable)
- `created_at`: DateTime (auto-generated)

**Validation Rules**:
- Event type must match operation context
- Message required and descriptive
- Metadata JSON must be valid if provided

**Relationships**:
- Many-to-one with FileTransfer
- Many-to-one with FileChunk (if chunk-specific)

## Data Flow & Business Logic

### Upload Flow
1. User initiates upload → FileTransfer created (status: pending)
2. File chunked → FileChunks created (status: pending)
3. Transfer starts → status: transferring, chunks sent sequentially
4. Progress tracked → completed_chunks updated
5. Final verification → status: completed, file reassembled

### Download Flow
1. User requests download → FileTransfer created (direction: download)
2. Remote file located → metadata populated
3. Chunks downloaded → FileChunks stored
4. Reassembly → local file created
5. Integrity check → transfer completed

### Resume Logic
1. Interrupted transfer detected → TransferSession checked
2. Unacknowledged chunks identified → resent from last good chunk
3. Progress continues from interruption point
4. No duplicate data transfer

### Cleanup Logic
1. Completed transfers → chunks deleted after retention period
2. Failed transfers → partial data cleaned up
3. Session expiry → stale sessions terminated
4. User deletion → all associated data removed

## Performance Considerations

### Chunk Size Optimization
- 1MB default chunks balance memory usage and transfer efficiency
- Transport-specific sizing (smaller for messaging limits)
- Dynamic adjustment based on network conditions

### Database Indexing
- Composite index on FileChunk (transfer_id, chunk_index)
- Index on FileTransfer (user_id, status, created_at)
- Index on TransferSession (session_token, expires_at)

### Query Patterns
- Transfer progress: COUNT completed_chunks / total_chunks
- Active transfers: WHERE status IN (transferring, reassembling)
- User quota: SUM file_size WHERE user_id = ? AND created_at > ?

## Security Model

### Access Control
- Users can only access their own transfers
- Transfer sessions bound to user authentication
- File data encrypted at rest and in transit

### Data Protection
- Chunk checksums prevent tampering
- Transport encryption protects in-flight data
- Optional E2E encryption for sensitive files

### Audit Trail
- All transfer events logged
- Failed attempts tracked for security monitoring
- Session activity monitored for abuse detection

## Migration Strategy

### Schema Evolution
- Add new tables without affecting existing functionality
- Backward compatibility for existing transfer mechanisms
- Gradual rollout with feature flags

### Data Migration
- Existing text-based transfers remain unchanged
- New binary transfers use extended schema
- Historical data preserved for audit purposes