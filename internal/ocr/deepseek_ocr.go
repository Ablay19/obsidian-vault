package ocr

import (
	"context"
	"encoding/json"
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/mathocr"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// DocumentLayout represents the structure of a document
type DocumentLayout struct {
	Title      string            `json:"title"`
	Sections   []DocumentSection `json:"sections"`
	HasTables  bool              `json:"has_tables"`
	HasImages  bool              `json:"has_images"`
	Language   string            `json:"language"`
	Confidence float64           `json:"confidence"`
}

// DocumentSection represents a section of the document
type DocumentSection struct {
	Type       string  `json:"type"` // title, paragraph, list, table, image
	Content    string  `json:"content"`
	Confidence float64 `json:"confidence"`
	Position   Rect    `json:"position"`
}

// Rect represents a rectangle position
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DeepSeekOCR implements advanced OCR using DeepSeek's document understanding
type DeepSeekOCR struct {
	aiService     ai.AIServiceInterface
	mathProcessor *mathocr.MathOCRProcessor
}

// NewDeepSeekOCR creates a new DeepSeek OCR processor
func NewDeepSeekOCR(aiService ai.AIServiceInterface) *DeepSeekOCR {
	return &DeepSeekOCR{
		aiService:     aiService,
		mathProcessor: mathocr.NewMathOCRProcessor(),
	}
}

// ProcessDocument performs advanced OCR and layout analysis on a document image
func (ocr *DeepSeekOCR) ProcessDocument(ctx context.Context, imagePath string) (*DocumentLayout, error) {
	if !ocr.isAvailable() {
		return nil, fmt.Errorf("DeepSeek OCR not available")
	}

	zap.S().Info("Starting DeepSeek document OCR processing", "path", imagePath)

	// Multi-stage processing inspired by DeepSeek-OCR
	layout, err := ocr.analyzeDocumentLayout(ctx, imagePath)
	if err != nil {
		return nil, fmt.Errorf("layout analysis failed: %w", err)
	}

	// Enhance text extraction with layout awareness
	layout, err = ocr.enhanceTextExtraction(ctx, imagePath, layout)
	if err != nil {
		zap.S().Warn("Text enhancement failed, continuing with basic layout", "error", err)
	}

	// Detect and extract tables
	layout, err = ocr.extractTables(ctx, imagePath, layout)
	if err != nil {
		zap.S().Warn("Table extraction failed", "error", err)
	}

	// Final confidence scoring
	layout.Confidence = ocr.calculateOverallConfidence(layout)

	zap.S().Info("DeepSeek OCR processing completed",
		"sections", len(layout.Sections),
		"has_tables", layout.HasTables,
		"confidence", layout.Confidence)

	return layout, nil
}

// ExtractText extracts clean text from document layout
func (ocr *DeepSeekOCR) ExtractText(layout *DocumentLayout) string {
	var text strings.Builder

	for i, section := range layout.Sections {
		switch section.Type {
		case "title":
			if i > 0 {
				text.WriteString("\n\n")
			}
			text.WriteString("# ")
			text.WriteString(section.Content)
			text.WriteString("\n")
		case "paragraph":
			if i > 0 {
				text.WriteString("\n\n")
			}
			text.WriteString(section.Content)
		case "list":
			text.WriteString("\n")
			text.WriteString(section.Content)
			text.WriteString("\n")
		case "table":
			text.WriteString("\n[Table]\n")
			text.WriteString(section.Content)
			text.WriteString("\n")
		default:
			text.WriteString(section.Content)
		}
	}

	return strings.TrimSpace(text.String())
}

// analyzeDocumentLayout performs initial layout analysis
func (ocr *DeepSeekOCR) analyzeDocumentLayout(ctx context.Context, imagePath string) (*DocumentLayout, error) {
	prompt := fmt.Sprintf(`Analyze this document image and provide a detailed layout analysis in JSON format.

Image: %s

Identify and extract:
1. Document title/main heading
2. Section headers and their hierarchy
3. Paragraph text blocks
4. Lists (numbered or bulleted)
5. Tables (if any)
6. Images or figures (if any)
7. Primary language of the document

Return a JSON object with:
{
  "title": "document title",
  "sections": [
    {
      "type": "title|paragraph|list|table|image",
      "content": "extracted content",
      "confidence": 0.0-1.0,
      "position": {"x": 0, "y": 0, "width": 100, "height": 50}
    }
  ],
  "has_tables": false,
  "has_images": false,
  "language": "english|french|etc"
}`, imagePath)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.1,
	}

	var response strings.Builder
	err := ocr.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var layout DocumentLayout
	jsonStr := strings.TrimSpace(response.String())
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimSuffix(jsonStr, "```")

	if err := json.Unmarshal([]byte(jsonStr), &layout); err != nil {
		return nil, fmt.Errorf("failed to parse layout JSON: %w", err)
	}

	return &layout, nil
}

// enhanceTextExtraction improves text extraction using layout information
func (ocr *DeepSeekOCR) enhanceTextExtraction(ctx context.Context, imagePath string, layout *DocumentLayout) (*DocumentLayout, error) {
	// Use layout information to guide more accurate OCR
	prompt := fmt.Sprintf(`Based on the document layout analysis, perform enhanced text extraction.

Layout summary:
- Title: %s
- Sections: %d
- Has tables: %t
- Language: %s

Image: %s

Extract clean, well-formatted text that preserves:
1. Proper paragraph breaks
2. List formatting
3. Table structure (if present)
4. Reading order
5. Special characters and formatting

Return the complete extracted text with proper formatting.`, layout.Title, len(layout.Sections), layout.HasTables, layout.Language, imagePath)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.1,
	}

	var response strings.Builder
	err := ocr.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return layout, err
	}

	// Update sections with enhanced text
	enhancedText := strings.TrimSpace(response.String())

	// Apply mathematical OCR enhancements
	enhancedText, formulas := ocr.mathProcessor.EnhanceOCROutput(enhancedText)

	// Log detected formulas
	if len(formulas) > 0 {
		zap.S().Info("Detected mathematical formulas", "count", len(formulas))
		for _, formula := range formulas {
			zap.S().Debug("Formula detected", "type", formula.Type, "text", formula.Text, "confidence", formula.Confidence)
		}
	}

	layout.Sections = ocr.parseEnhancedTextIntoSections(enhancedText)

	return layout, nil
}

// extractTables performs specialized table extraction
func (ocr *DeepSeekOCR) extractTables(ctx context.Context, imagePath string, layout *DocumentLayout) (*DocumentLayout, error) {
	if !layout.HasTables {
		return layout, nil
	}

	prompt := fmt.Sprintf(`Extract and format all tables from this document image as clean text.

Image: %s

For each table found:
1. Identify table structure (rows, columns)
2. Extract cell contents accurately
3. Format as markdown table or structured text
4. Preserve numerical data and formatting

Return all tables with clear labels and formatting.`, imagePath)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.1,
	}

	var response strings.Builder
	err := ocr.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return layout, err
	}

	// Add table sections to layout
	tableContent := strings.TrimSpace(response.String())
	if tableContent != "" {
		layout.Sections = append(layout.Sections, DocumentSection{
			Type:       "table",
			Content:    tableContent,
			Confidence: 0.8,
		})
	}

	return layout, nil
}

// parseEnhancedTextIntoSections converts enhanced text back into structured sections
func (ocr *DeepSeekOCR) parseEnhancedTextIntoSections(text string) []DocumentSection {
	var sections []DocumentSection

	// Split by double newlines to identify major sections
	parts := strings.Split(text, "\n\n")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		section := DocumentSection{
			Content:    part,
			Confidence: 0.9,
		}

		// Classify section type
		if strings.HasPrefix(part, "# ") {
			section.Type = "title"
			section.Content = strings.TrimPrefix(part, "# ")
		} else if ocr.isList(part) {
			section.Type = "list"
		} else if ocr.isTable(part) {
			section.Type = "table"
		} else {
			section.Type = "paragraph"
		}

		sections = append(sections, section)
	}

	return sections
}

// isList checks if text represents a list
func (ocr *DeepSeekOCR) isList(text string) bool {
	lines := strings.Split(text, "\n")
	listIndicators := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for numbered lists (1., 2., etc.)
		if matched, _ := regexp.MatchString(`^\d+\.`, line); matched {
			listIndicators++
		}

		// Check for bulleted lists (-, *, •)
		if matched, _ := regexp.MatchString(`^[-*•]\s`, line); matched {
			listIndicators++
		}
	}

	return listIndicators > 0 && float64(listIndicators)/float64(len(lines)) > 0.5
}

// isTable checks if text represents a table
func (ocr *DeepSeekOCR) isTable(text string) bool {
	lines := strings.Split(text, "\n")
	pipeCount := 0

	for _, line := range lines {
		if strings.Contains(line, "|") {
			pipeCount++
		}
	}

	return pipeCount > 1 && float64(pipeCount)/float64(len(lines)) > 0.6
}

// calculateOverallConfidence computes document-level confidence
func (ocr *DeepSeekOCR) calculateOverallConfidence(layout *DocumentLayout) float64 {
	if len(layout.Sections) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, section := range layout.Sections {
		totalConfidence += section.Confidence
	}

	avgConfidence := totalConfidence / float64(len(layout.Sections))

	// Boost confidence for well-structured documents
	structureBonus := 0.0
	if layout.Title != "" {
		structureBonus += 0.1
	}
	if len(layout.Sections) > 3 {
		structureBonus += 0.1
	}
	if layout.HasTables {
		structureBonus += 0.05
	}

	confidence := avgConfidence + structureBonus
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// isAvailable checks if DeepSeek OCR is available
func (ocr *DeepSeekOCR) isAvailable() bool {
	return ocr.aiService != nil &&
		strings.Contains(strings.ToLower(ocr.aiService.GetActiveProviderName()), "deepseek")
}
