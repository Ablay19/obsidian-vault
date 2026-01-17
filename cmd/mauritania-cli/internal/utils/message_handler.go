package utils

import (
	"fmt"
	"strings"
	"time"
)

// MessageHandler handles message size limits and pagination for social media transports
type MessageHandler struct {
	maxMessageLength int
	chunkPrefix      string
	chunkSeparator   string
	logger           *Logger
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(logger *Logger) *MessageHandler {
	return &MessageHandler{
		maxMessageLength: 4000, // Conservative limit for most social media platforms
		chunkPrefix:      "[Part %d/%d] ",
		chunkSeparator:   "\n...\n",
		logger:           logger,
	}
}

// SplitMessage splits a long message into chunks that fit within platform limits
func (mh *MessageHandler) SplitMessage(message string) []string {
	if len(message) <= mh.maxMessageLength {
		return []string{message}
	}

	var chunks []string
	remaining := message

	for i := 1; len(remaining) > 0; i++ {
		// Calculate how much content we can fit in this chunk
		prefix := fmt.Sprintf(mh.chunkPrefix, i, 0) // We'll update the total later
		availableLength := mh.maxMessageLength - len(prefix) - len(mh.chunkSeparator)

		if availableLength <= 0 {
			// Message is too long even for the prefix, truncate
			chunk := remaining
			if len(chunk) > mh.maxMessageLength {
				chunk = chunk[:mh.maxMessageLength-3] + "..."
			}
			chunks = append(chunks, chunk)
			break
		}

		// Try to split at word boundaries
		chunk := remaining
		if len(remaining) > availableLength {
			chunk = remaining[:availableLength]

			// Try to find a good break point
			if lastSpace := strings.LastIndex(chunk, " "); lastSpace > availableLength/2 {
				chunk = chunk[:lastSpace]
			}

			remaining = remaining[len(chunk):]
		} else {
			remaining = ""
		}

		// Add the chunk
		chunks = append(chunks, chunk)
	}

	// Update chunk prefixes with correct total count
	totalChunks := len(chunks)
	for i, chunk := range chunks {
		if totalChunks > 1 {
			prefix := fmt.Sprintf(mh.chunkPrefix, i+1, totalChunks)
			chunks[i] = prefix + chunk
		}
	}

	return chunks
}

// CombineMessage combines paginated message chunks back into a single message
func (mh *MessageHandler) CombineMessage(chunks []string) string {
	if len(chunks) == 0 {
		return ""
	}

	if len(chunks) == 1 {
		return mh.stripChunkPrefix(chunks[0])
	}

	var combined strings.Builder
	for _, chunk := range chunks {
		cleanChunk := mh.stripChunkPrefix(chunk)
		combined.WriteString(cleanChunk)
		combined.WriteString(mh.chunkSeparator)
	}

	return strings.TrimSuffix(combined.String(), mh.chunkSeparator)
}

// IsChunkedMessage checks if a message is part of a chunked sequence
func (mh *MessageHandler) IsChunkedMessage(message string) bool {
	return strings.Contains(message, "[Part ") && strings.Contains(message, "/")
}

// ExtractChunkInfo extracts chunk information from a chunked message
func (mh *MessageHandler) ExtractChunkInfo(message string) (current, total int, content string, ok bool) {
	if !mh.IsChunkedMessage(message) {
		return 0, 0, message, false
	}

	// Find the prefix pattern [Part X/Y]
	start := strings.Index(message, "[Part ")
	if start == -1 {
		return 0, 0, message, false
	}

	end := strings.Index(message[start:], "]")
	if end == -1 {
		return 0, 0, message, false
	}

	prefix := message[start : start+end+1]
	content = strings.TrimPrefix(message, prefix)

	// Parse "Part X/Y"
	parts := strings.Split(strings.Trim(prefix, "[]"), " ")
	if len(parts) != 2 {
		return 0, 0, message, false
	}

	chunkParts := strings.Split(parts[1], "/")
	if len(chunkParts) != 2 {
		return 0, 0, message, false
	}

	if currentPart, err := parseInt(chunkParts[0]); err == nil {
		if totalPart, err := parseInt(chunkParts[1]); err == nil {
			return currentPart, totalPart, content, true
		}
	}

	return 0, 0, message, false
}

// stripChunkPrefix removes the chunk prefix from a message
func (mh *MessageHandler) stripChunkPrefix(message string) string {
	if !mh.IsChunkedMessage(message) {
		return message
	}

	start := strings.Index(message, "[Part ")
	if start == -1 {
		return message
	}

	end := strings.Index(message[start:], "]")
	if end == -1 {
		return message
	}

	return message[start+end+2:] // +2 for "] "
}

// TruncateMessage truncates a message to fit within limits with ellipsis
func (mh *MessageHandler) TruncateMessage(message string) string {
	if len(message) <= mh.maxMessageLength {
		return message
	}

	truncated := message[:mh.maxMessageLength-3] + "..."
	return truncated
}

// ValidateMessageLength validates that a message fits within platform limits
func (mh *MessageHandler) ValidateMessageLength(message string) error {
	if len(message) > mh.maxMessageLength {
		return fmt.Errorf("message length %d exceeds maximum allowed length %d", len(message), mh.maxMessageLength)
	}
	return nil
}

// EstimateChunks estimates how many chunks a message will be split into
func (mh *MessageHandler) EstimateChunks(message string) int {
	if len(message) <= mh.maxMessageLength {
		return 1
	}

	// Rough estimation - actual splitting is more sophisticated
	avgChunkSize := mh.maxMessageLength - len(fmt.Sprintf(mh.chunkPrefix, 1, 10)) - len(mh.chunkSeparator)
	if avgChunkSize <= 0 {
		return 1
	}

	return (len(message) + avgChunkSize - 1) / avgChunkSize // Ceiling division
}

// MessageQueue manages message chunks for ordered delivery
type MessageQueue struct {
	chunks  map[string][]*QueuedMessageChunk
	logger  *Logger
	timeout time.Duration
}

// QueuedMessageChunk represents a chunk of a multi-part message
type QueuedMessageChunk struct {
	MessageID   string
	ChunkIndex  int
	TotalChunks int
	Content     string
	Timestamp   time.Time
	SenderID    string
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(logger *Logger) *MessageQueue {
	return &MessageQueue{
		chunks:  make(map[string][]*QueuedMessageChunk),
		logger:  logger,
		timeout: 5 * time.Minute, // 5 minutes to collect all chunks
	}
}

// AddChunk adds a message chunk to the queue
func (mq *MessageQueue) AddChunk(senderID string, message string) ([]string, bool) {
	chunk := &QueuedMessageChunk{
		SenderID:  senderID,
		Timestamp: time.Now(),
	}

	current, total, content, isChunked := NewMessageHandler(mq.logger).ExtractChunkInfo(message)

	if !isChunked {
		// Single message, return as-is
		return []string{content}, true
	}

	chunk.MessageID = fmt.Sprintf("%s_%d", senderID, current)
	chunk.ChunkIndex = current
	chunk.TotalChunks = total
	chunk.Content = content

	// Add to queue
	messageKey := fmt.Sprintf("%s_%s", senderID, time.Now().Format("20060102_150405"))
	if mq.chunks[messageKey] == nil {
		mq.chunks[messageKey] = make([]*QueuedMessageChunk, 0, total)
	}

	mq.chunks[messageKey] = append(mq.chunks[messageKey], chunk)

	// Check if we have all chunks
	chunks := mq.chunks[messageKey]
	if len(chunks) == total {
		// Sort by chunk index and combine
		sortedChunks := make([]string, total)
		for _, c := range chunks {
			if c.ChunkIndex <= total && c.ChunkIndex > 0 {
				sortedChunks[c.ChunkIndex-1] = c.Content
			}
		}

		// Clean up
		delete(mq.chunks, messageKey)

		return sortedChunks, true
	}

	return nil, false // Not complete yet
}

// Cleanup removes old incomplete message chunks
func (mq *MessageQueue) Cleanup() {
	now := time.Now()
	for key, chunks := range mq.chunks {
		if len(chunks) > 0 && now.Sub(chunks[0].Timestamp) > mq.timeout {
			mq.logger.Info("Cleaning up incomplete message: %s", key)
			delete(mq.chunks, key)
		}
	}
}

// parseInt is a simple integer parser (stdlib strconv would be better but avoiding imports)
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid digit")
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
