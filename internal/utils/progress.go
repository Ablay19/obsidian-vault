package utils

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// ProgressBar represents a progress bar with timing information
type ProgressBar struct {
	total       int
	current     int
	width       int
	startTime   time.Time
	lastUpdate  time.Time
	showTime    bool
	showPercent bool
	description string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, description string) *ProgressBar {
	return &ProgressBar{
		total:       total,
		current:     0,
		width:       20,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
		showTime:    true,
		showPercent: true,
		description: description,
	}
}

// SetWidth sets the width of the progress bar
func (pb *ProgressBar) SetWidth(width int) *ProgressBar {
	pb.width = width
	return pb
}

// SetShowTime sets whether to show time estimates
func (pb *ProgressBar) SetShowTime(show bool) *ProgressBar {
	pb.showTime = show
	return pb
}

// SetShowPercent sets whether to show percentage
func (pb *ProgressBar) SetShowPercent(show bool) *ProgressBar {
	pb.showPercent = show
	return pb
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int) {
	pb.current = current
	pb.lastUpdate = time.Now()
}

// Increment increments the progress bar by 1
func (pb *ProgressBar) Increment() {
	pb.Update(pb.current + 1)
}

// SetDescription sets the description
func (pb *ProgressBar) SetDescription(desc string) {
	pb.description = desc
}

// GetProgress returns current progress as a percentage (0-100)
func (pb *ProgressBar) GetProgress() float64 {
	if pb.total == 0 {
		return 100.0
	}
	return math.Min(100.0, (float64(pb.current)/float64(pb.total))*100.0)
}

// GetETA returns estimated time to completion in seconds
func (pb *ProgressBar) GetETA() float64 {
	if pb.current == 0 {
		return 0
	}

	elapsed := time.Since(pb.startTime).Seconds()
	progress := float64(pb.current) / float64(pb.total)

	if progress >= 1.0 {
		return 0
	}

	totalEstimated := elapsed / progress
	return totalEstimated - elapsed
}

// Render renders the progress bar as a string (npm/yarn style)
func (pb *ProgressBar) Render() string {
	progress := pb.GetProgress()
	filled := int(math.Round(float64(pb.width) * progress / 100.0))

	// Create progress bar (npm/yarn style with ▌ and ▐)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)

	// Build status string in npm/yarn style
	var parts []string

	if pb.description != "" {
		parts = append(parts, pb.description)
	}

	parts = append(parts, fmt.Sprintf("[%s] %d/%d", bar, pb.current, pb.total))

	if pb.showPercent {
		parts = append(parts, fmt.Sprintf("%.0f%%", progress))
	}

	if pb.showTime && pb.current > 0 {
		eta := pb.GetETA()
		if eta > 0 {
			if eta < 60 {
				parts = append(parts, fmt.Sprintf("(%ds)", int(eta)))
			} else if eta < 3600 {
				parts = append(parts, fmt.Sprintf("(%.0fm)", eta/60))
			} else {
				parts = append(parts, fmt.Sprintf("(%.1fh)", eta/3600))
			}
		}
	}

	return strings.Join(parts, " ")
}

// IsComplete returns true if the progress bar is complete
func (pb *ProgressBar) IsComplete() bool {
	return pb.current >= pb.total
}

// Complete marks the progress bar as complete
func (pb *ProgressBar) Complete() {
	pb.Update(pb.total)
}

// GetStartTime returns the start time of the progress bar
func (pb *ProgressBar) GetStartTime() time.Time {
	return pb.startTime
}

// ProgressTracker manages multiple progress bars for complex operations
type ProgressTracker struct {
	bars    map[string]*ProgressBar
	order   []string
	current string
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		bars:  make(map[string]*ProgressBar),
		order: []string{},
	}
}

// AddBar adds a progress bar with the given name and total
func (pt *ProgressTracker) AddBar(name, description string, total int) *ProgressBar {
	bar := NewProgressBar(total, description)
	pt.bars[name] = bar
	pt.order = append(pt.order, name)
	return bar
}

// SetCurrent sets the currently active progress bar
func (pt *ProgressTracker) SetCurrent(name string) {
	pt.current = name
}

// GetCurrent returns the currently active progress bar
func (pt *ProgressTracker) GetCurrent() *ProgressBar {
	if pt.current == "" && len(pt.order) > 0 {
		return pt.bars[pt.order[0]]
	}
	return pt.bars[pt.current]
}

// GetBar returns a specific progress bar by name
func (pt *ProgressTracker) GetBar(name string) *ProgressBar {
	return pt.bars[name]
}

// RenderCurrent renders the current progress bar
func (pt *ProgressTracker) RenderCurrent() string {
	if bar := pt.GetCurrent(); bar != nil {
		return bar.Render()
	}
	return ""
}

// RenderAll renders all progress bars
func (pt *ProgressTracker) RenderAll() string {
	var parts []string
	for _, name := range pt.order {
		if bar := pt.bars[name]; bar != nil {
			parts = append(parts, bar.Render())
		}
	}
	return strings.Join(parts, "\n")
}

// IsComplete checks if all progress bars are complete
func (pt *ProgressTracker) IsComplete() bool {
	for _, bar := range pt.bars {
		if !bar.IsComplete() {
			return false
		}
	}
	return true
}

// ProcessingStages defines the stages of image processing
var ProcessingStages = []string{
	"upload",
	"validation",
	"ocr_extraction",
	"vision_encoding",
	"multimodal_fusion",
	"ai_analysis",
	"summarization",
	"storage",
}

// CreateImageProcessingTracker creates a progress tracker for image processing
func CreateImageProcessingTracker() *ProgressTracker {
	tracker := NewProgressTracker()

	// Define stage totals and descriptions (professional style)
	stages := map[string]struct {
		total       int
		description string
	}{
		"upload": {
			total:       1,
			description: "Receiving image",
		},
		"validation": {
			total:       1,
			description: "Validating image",
		},
		"ocr_extraction": {
			total:       100, // Percentage-based
			description: "Extracting text",
		},
		"vision_encoding": {
			total:       100, // Percentage-based
			description: "Vision encoding",
		},
		"multimodal_fusion": {
			total:       100, // Percentage-based
			description: "Multimodal fusion",
		},
		"ai_analysis": {
			total:       100, // Percentage-based
			description: "AI analysis",
		},
		"summarization": {
			total:       1,
			description: "Generating summary",
		},
		"storage": {
			total:       1,
			description: "Saving results",
		},
	}

	for _, stage := range ProcessingStages {
		if info, exists := stages[stage]; exists {
			tracker.AddBar(stage, info.description, info.total)
		}
	}

	return tracker
}
