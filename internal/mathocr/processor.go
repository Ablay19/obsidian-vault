package mathocr

import (
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// MathFormula represents a detected mathematical formula
type MathFormula struct {
	Text       string  `json:"text"`
	LaTeX      string  `json:"latex,omitempty"`
	Position   Rect    `json:"position"`
	Confidence float64 `json:"confidence"`
	Type       string  `json:"type"` // limit, integral, derivative, etc.
}

// Rect represents a position rectangle
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// MathOCRProcessor enhances OCR for mathematical content
type MathOCRProcessor struct {
	enableLaTeXConversion  bool
	enableFormulaDetection bool
	minFormulaConfidence   float64
}

// NewMathOCRProcessor creates a new mathematical OCR processor
func NewMathOCRProcessor() *MathOCRProcessor {
	return &MathOCRProcessor{
		enableLaTeXConversion:  true,
		enableFormulaDetection: true,
		minFormulaConfidence:   0.7,
	}
}

// ProcessMathematicalText enhances text extraction for mathematical content
func (m *MathOCRProcessor) ProcessMathematicalText(text string) string {
	if !m.enableFormulaDetection {
		return text
	}

	zap.S().Info("Processing text for mathematical content enhancement")

	// Apply mathematical text enhancements
	enhancedText := m.normalizeMathematicalSymbols(text)
	enhancedText = m.fixCommonMathOCRErrors(enhancedText)
	enhancedText = m.reconstructMathematicalExpressions(enhancedText)
	enhancedText = m.formatMathematicalNotation(enhancedText)

	return enhancedText
}

// DetectMathematicalFormulas identifies mathematical formulas in text
func (m *MathOCRProcessor) DetectMathematicalFormulas(text string) []MathFormula {
	var formulas []MathFormula

	if !m.enableFormulaDetection {
		return formulas
	}

	// Detect limits
	limitPatterns := []string{
		`lim\s*\([^)]+\)`,
		`lim\s*_\s*\{[^}]+\}`,
		`\\lim`,
	}

	for _, pattern := range limitPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		for _, match := range matches {
			if confidence := m.calculateFormulaConfidence(match); confidence >= m.minFormulaConfidence {
				formulas = append(formulas, MathFormula{
					Text:       match,
					Type:       "limit",
					Confidence: confidence,
				})
			}
		}
	}

	// Detect integrals
	integralPatterns := []string{
		`\\int`,
		`∫`,
		`int\s*\([^)]+\)`,
	}

	for _, pattern := range integralPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		for _, match := range matches {
			if confidence := m.calculateFormulaConfidence(match); confidence >= m.minFormulaConfidence {
				formulas = append(formulas, MathFormula{
					Text:       match,
					Type:       "integral",
					Confidence: confidence,
				})
			}
		}
	}

	// Detect derivatives
	derivativePatterns := []string{
		`d/dx`,
		`\\frac\{d\}\{dx\}`,
		`f'\([^)]+\)`,
		`\\frac\{d\^?\d*\}\{d[^}]+\}`,
	}

	for _, pattern := range derivativePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		for _, match := range matches {
			if confidence := m.calculateFormulaConfidence(match); confidence >= m.minFormulaConfidence {
				formulas = append(formulas, MathFormula{
					Text:       match,
					Type:       "derivative",
					Confidence: confidence,
				})
			}
		}
	}

	// Detect summations and products
	sumProdPatterns := []string{
		`\\sum`,
		`\\prod`,
		`Σ`,
		`Π`,
	}

	for _, pattern := range sumProdPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		for _, match := range matches {
			if confidence := m.calculateFormulaConfidence(match); confidence >= m.minFormulaConfidence {
				formulas = append(formulas, MathFormula{
					Text:       match,
					Type:       "summation_product",
					Confidence: confidence,
				})
			}
		}
	}

	return formulas
}

// ConvertToLaTeX attempts to convert mathematical expressions to LaTeX
func (m *MathOCRProcessor) ConvertToLaTeX(text string) string {
	if !m.enableLaTeXConversion {
		return text
	}

	// Basic LaTeX conversions
	conversions := map[string]string{
		"lim":     "\\lim",
		"lim (":   "\\lim_{",
		"lim(":    "\\lim_{",
		"->":      "\\to",
		"→":       "\\to",
		"∞":       "\\infty",
		"alpha":   "\\alpha",
		"beta":    "\\beta",
		"gamma":   "\\gamma",
		"delta":   "\\delta",
		"epsilon": "\\epsilon",
		"theta":   "\\theta",
		"lambda":  "\\lambda",
		"mu":      "\\mu",
		"pi":      "\\pi",
		"sigma":   "\\sigma",
		"phi":     "\\phi",
		"omega":   "\\omega",
		"sqrt(":   "\\sqrt{",
		"sin(":    "\\sin(",
		"cos(":    "\\cos(",
		"tan(":    "\\tan(",
		"log(":    "\\log(",
		"ln(":     "\\ln(",
		"exp(":    "\\exp(",
		"sum":     "\\sum",
		"prod":    "\\prod",
		"int":     "\\int",
		"frac":    "\\frac",
		"sqrt":    "\\sqrt",
		"times":   "\\times",
		"div":     "\\div",
		"pm":      "\\pm",
		"mp":      "\\mp",
		"leq":     "\\leq",
		"geq":     "\\geq",
		"neq":     "\\neq",
		"approx":  "\\approx",
		"equiv":   "\\equiv",
	}

	result := text
	for ocrText, latex := range conversions {
		result = strings.ReplaceAll(result, ocrText, latex)
	}

	// Handle fractions like a/b
	result = m.convertSimpleFractions(result)

	// Handle superscripts and subscripts
	result = m.convertSuperSubScripts(result)

	return result
}

// normalizeMathematicalSymbols fixes common OCR errors in mathematical symbols
func (m *MathOCRProcessor) normalizeMathematicalSymbols(text string) string {
	// Common OCR misrecognitions for mathematical symbols
	replacements := map[string]string{
		"lim":   "lim",
		"|im":   "lim",
		"!im":   "lim",
		"l!m":   "lim",
		"ln(":   "ln(",
		"1n(":   "ln(",
		"exp(":  "exp(",
		"ex p(": "exp(",
		"e xp(": "exp(",
		"sin(":  "sin(",
		"s!n(":  "sin(",
		"cos(":  "cos(",
		"c0s(":  "cos(",
		"tan(":  "tan(",
		"t4n(":  "tan(",
		"->":    "→",
		"~>":    "→",
		"—>":    "→",
		"∞":     "∞",
		"oo":    "∞",
		"00":    "∞",
		"inf":   "∞",
		"+-":    "±",
		"-+":    "∓",
		"<=":    "≤",
		">=":    "≥",
		"! =":   "≠",
		"!=":    "≠",
		"~=":    "≈",
		"==":    "≡",
		"sqrt(": "√(",
		"∫":     "∫",
		"Σ":     "Σ",
		"Π":     "Π",
		"Δ":     "Δ",
		"α":     "α",
		"β":     "β",
		"γ":     "γ",
		"δ":     "δ",
		"λ":     "λ",
		"μ":     "μ",
		"π":     "π",
		"σ":     "σ",
		"φ":     "φ",
		"ω":     "ω",
	}

	result := text
	for ocr, correct := range replacements {
		result = strings.ReplaceAll(result, ocr, correct)
	}

	return result
}

// fixCommonMathOCRErrors fixes common OCR errors specific to mathematical content
func (m *MathOCRProcessor) fixCommonMathOCRErrors(text string) string {
	// Fix spacing issues in mathematical expressions
	text = regexp.MustCompile(`(\w)\s*(\+|×|÷|−)\s*(\w)`).ReplaceAllString(text, "$1 $2 $3")

	// Fix parentheses spacing
	text = regexp.MustCompile(`(\w)\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(text, "$1($2)")

	// Fix function call spacing
	text = regexp.MustCompile(`(\w+)\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(text, "$1($2)")

	// Fix exponent notation
	text = regexp.MustCompile(`(\w+)\s*\^\s*(\w+)`).ReplaceAllString(text, "$1^$2")

	// Fix subscript notation
	text = regexp.MustCompile(`(\w+)\s*_\s*(\w+)`).ReplaceAllString(text, "$1_$2")

	return text
}

// reconstructMathematicalExpressions attempts to reconstruct broken mathematical expressions
func (m *MathOCRProcessor) reconstructMathematicalExpressions(text string) string {
	// Fix broken limit expressions
	text = regexp.MustCompile(`lim\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(text, "lim_{$1}")

	// Fix broken fraction expressions
	text = regexp.MustCompile(`(\d+)/(\d+)`).ReplaceAllString(text, "\\frac{$1}{$2}")

	// Fix broken integral expressions
	text = regexp.MustCompile(`int\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(text, "\\int_{$1}")

	// Fix broken derivative expressions
	text = regexp.MustCompile(`d/dx\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(text, "\\frac{d}{dx}($1)")

	return text
}

// formatMathematicalNotation applies consistent formatting to mathematical expressions
func (m *MathOCRProcessor) formatMathematicalNotation(text string) string {
	// Add spaces around operators for readability
	text = regexp.MustCompile(`([^+\-×÷=≠≈≡≤≥])\s*([+\-×÷=≠≈≡≤≥])\s*([^+\-×÷=≠≈≡≤≥])`).ReplaceAllString(text, "$1 $2 $3")

	// Format function calls consistently
	text = regexp.MustCompile(`(\w+)\s*\(\s*([^)]+)\s*\)`).ReplaceAllString(text, "$1($2)")

	return text
}

// convertSimpleFractions converts a/b style fractions to LaTeX
func (m *MathOCRProcessor) convertSimpleFractions(text string) string {
	// Convert simple fractions like 1/2 to \frac{1}{2}
	re := regexp.MustCompile(`(\d+)/(\d+)`)
	return re.ReplaceAllString(text, "\\frac{$1}{$2}")
}

// convertSuperSubScripts attempts to detect and format superscripts and subscripts
func (m *MathOCRProcessor) convertSuperSubScripts(text string) string {
	// This is a simplified implementation - in practice, this would need
	// more sophisticated pattern matching based on font size and position
	// For now, we'll handle common cases like x^2 and x_1

	// Handle superscripts (simplified)
	text = regexp.MustCompile(`(\w+)\s*\^\s*(\d+|[a-zA-Z])`).ReplaceAllString(text, "$1^{$2}")

	// Handle subscripts (simplified)
	text = regexp.MustCompile(`(\w+)\s*_\s*(\d+|[a-zA-Z])`).ReplaceAllString(text, "$1_{$2}")

	return text
}

// calculateFormulaConfidence calculates confidence score for a detected formula
func (m *MathOCRProcessor) calculateFormulaConfidence(formula string) float64 {
	// Simple confidence calculation based on formula characteristics
	confidence := 0.5 // Base confidence

	// Increase confidence for recognized mathematical symbols
	mathSymbols := []string{"lim", "int", "sum", "prod", "frac", "sqrt", "sin", "cos", "tan", "log", "ln", "exp"}
	for _, symbol := range mathSymbols {
		if strings.Contains(formula, symbol) {
			confidence += 0.1
		}
	}

	// Increase confidence for proper mathematical notation
	if strings.Contains(formula, "{") && strings.Contains(formula, "}") {
		confidence += 0.2
	}

	if strings.Contains(formula, "(") && strings.Contains(formula, ")") {
		confidence += 0.1
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// EnhanceOCROutput applies all mathematical enhancements to OCR output
func (m *MathOCRProcessor) EnhanceOCROutput(text string) (string, []MathFormula) {
	// Process text for mathematical content
	enhancedText := m.ProcessMathematicalText(text)

	// Detect formulas
	formulas := m.DetectMathematicalFormulas(enhancedText)

	// Convert to LaTeX if enabled
	if m.enableLaTeXConversion {
		enhancedText = m.ConvertToLaTeX(enhancedText)
	}

	return enhancedText, formulas
}
