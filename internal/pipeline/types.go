package pipeline

import (
	"context"
	"time"
)

// ContentType defines supported data formats.
type ContentType string

const (
	ContentTypeImage ContentType = "image"
	ContentTypePDF   ContentType = "pdf"
	ContentTypeText  ContentType = "text"
)

// String returns the string representation of the ContentType.
func (ct ContentType) String() string {
	return string(ct)
}

// Job represents a single ingestion task.
type Job struct {
	ID            string
	Source        string      // e.g., "telegram", "gdrive"
	SourceID      string      // e.g., Message ID or File ID
	ContentType   ContentType
	Data          []byte      // Or a path/URL
	FileLocalPath string      // New field: Local path to the downloaded file
	Metadata      map[string]interface{}
	ReceivedAt    time.Time
	RetryCount    int
	MaxRetries    int
	UserContext   UserContext // Info about the user (auth, preferences)
	OutputFormat  string      // e.g., "md", "pdf"
	GitCommit     bool        // Whether to commit/push
	ProcessedContent interface{} // New field: To store the processed content result
}

type UserContext struct {
	UserID    string
	GoogleID  string
	Language  string
}

// Result represents the output of a processing job.
type Result struct {
	JobID       string
	Success     bool
	Error       error
	ProcessedAt time.Time
	Output      interface{} // The processed note/content
}

// SourceConnector defines how data is fetched.
// (In reality, the Bot pushes to the pipeline, but this interface helps for polling sources like GDrive)
type SourceConnector interface {
	Name() string
	Start(ctx context.Context, jobChan chan<- Job) error
}

// Processor defines the transformation logic (OCR -> AI).
type Processor interface {
	Process(ctx context.Context, job Job) (Result, error)
}

// Sink defines where the result goes (Obsidian, Database).
type Sink interface {
	Save(ctx context.Context, job Job, result Result) error
}