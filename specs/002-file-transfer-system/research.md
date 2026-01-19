# Secure File Transfer System Research

## File Chunking Strategies

**Decision**: Use fixed 1MB chunks with adaptive sizing for transport limits
**Rationale**: 1MB provides good balance between memory usage and transfer efficiency. Adaptive sizing ensures compatibility with transport message limits (e.g., WhatsApp's 100MB limit).
**Alternatives Considered**: Dynamic chunking based on file size (complexity), fixed small chunks (overhead), variable chunks (consistency issues)

## Integrity Verification Methods

**Decision**: SHA-256 per chunk plus overall file hash, with Merkle tree for large files
**Rationale**: SHA-256 provides strong integrity guarantees. Per-chunk verification enables resumable transfers. Merkle tree allows efficient verification of large files without full re-transfer.
**Alternatives Considered**: CRC32 (weak security), MD5 (deprecated), full file hash only (no resume capability)

## Compression Algorithms

**Decision**: gzip for general files, zstd for large datasets, no compression for already compressed formats
**Rationale**: gzip provides good compression ratio with low CPU usage. zstd offers better performance for large files. Skipping compression for images/videos avoids wasted CPU.
**Alternatives Considered**: bzip2 (slower), LZ4 (less compression), custom algorithms (maintenance burden)

## Database Storage Options

**Decision**: SQLite with blob storage for chunks, metadata in structured tables
**Rationale**: SQLite provides ACID transactions and reliable storage. Blob storage handles binary data efficiently. Structured metadata enables fast queries and indexing.
**Alternatives Considered**: File system storage (no transactions), PostgreSQL (heavier), in-memory (not persistent)

## Transport Binary Encoding

**Decision**: Base64 for text-based transports, raw binary for others, with transport-specific optimizations
**Rationale**: Base64 ensures compatibility with text-only transports like messaging APIs. Raw binary maximizes efficiency for capable transports. Transport-specific handling accounts for different limitations.
**Alternatives Considered**: Hex encoding (inefficient), custom binary protocols (complex), compression-only (compatibility issues)

## Transfer Resume Mechanisms

**Decision**: Chunk-level resume with server-side state tracking and client-side progress persistence
**Rationale**: Chunk-level allows resuming from any interruption point. Server state ensures reliability. Client persistence handles app restarts.
**Alternatives Considered**: File-level resume only (less granular), no resume (poor UX), full re-transfer (wasteful)

## Progress Reporting

**Decision**: Real-time progress via existing CLI status mechanisms, with WebSocket support for future UI
**Rationale**: Leverages existing infrastructure for immediate compatibility. WebSocket provides foundation for real-time UI updates.
**Alternatives Considered**: Polling only (latency), push notifications (complex), no progress (poor UX)

## Error Handling and Recovery

**Decision**: Exponential backoff retry with circuit breaker pattern and user notifications
**Rationale**: Exponential backoff prevents overwhelming failing services. Circuit breaker provides resilience. User notifications maintain transparency.
**Alternatives Considered**: Fixed retry intervals (inefficient), immediate failure (poor reliability), silent retries (lack of feedback)

## Security Considerations

**Decision**: Transport-level encryption with optional end-to-end encryption for sensitive files
**Rationale**: Transport encryption leverages existing secure channels. Optional E2E provides additional security for sensitive data without overhead for all files.
**Alternatives Considered**: Always E2E (performance impact), no encryption (security risk), custom encryption (complexity)

## Performance Optimization

**Decision**: Concurrent chunk transfers with bandwidth throttling and priority queuing
**Rationale**: Concurrency maximizes throughput on reliable connections. Throttling prevents overwhelming transports. Priority queuing ensures critical transfers complete first.
**Alternatives Considered**: Sequential transfers (slower), unlimited concurrency (unstable), no queuing (fairness issues)