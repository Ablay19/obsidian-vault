package pipeline

import (
	"fmt"
)

// ValidateJob checks if a job meets strict validity criteria.
func ValidateJob(job Job) error {
	if job.ID == "" {
		return fmt.Errorf("missing job ID")
	}
	if job.Data == nil && job.ContentType != ContentTypeText { // Assuming generic checks
		return fmt.Errorf("job has no data")
	}

	switch job.ContentType {
	case ContentTypeImage, ContentTypePDF, ContentTypeText:
		// OK
	default:
		return fmt.Errorf("unsupported content type: %s", job.ContentType)
	}

	// Size Check (Simple Example)
	if len(job.Data) > 50*1024*1024 { // 50MB
		return fmt.Errorf("file size too large (%d bytes)", len(job.Data))
	}

	return nil
}
