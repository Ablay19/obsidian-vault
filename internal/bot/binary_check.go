package bot

import (
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// BinaryDependency represents an external binary dependency
type BinaryDependency struct {
	Name        string
	CheckCmd    []string
	InstallHelp string
}

// RequiredBinaries lists all external binary dependencies
var RequiredBinaries = []BinaryDependency{
	{
		Name:     "tesseract",
		CheckCmd: []string{"tesseract", "--version"},
		InstallHelp: "Install Tesseract OCR: https://tesseract-ocr.github.io/tessdoc/Installation.html\n" +
			"Ubuntu/Debian: sudo apt-get install tesseract-ocr\n" +
			"macOS: brew install tesseract",
	},
	{
		Name:     "pdftotext",
		CheckCmd: []string{"pdftotext", "-v"},
		InstallHelp: "Install pdftotext: https://poppler.freedesktop.org/\n" +
			"Ubuntu/Debian: sudo apt-get install poppler-utils\n" +
			"macOS: brew install poppler",
	},
}

// CheckBinary checks if a binary exists and is executable
func CheckBinary(dep BinaryDependency) error {
	cmd := exec.Command(dep.CheckCmd[0], dep.CheckCmd[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		zap.S().Errorw("Binary check failed",
			"binary", dep.Name,
			"error", err,
			"output", string(output),
		)
		return fmt.Errorf("binary %q not found or not executable", dep.Name)
	}

	zap.S().Infow("Binary check passed",
		"binary", dep.Name,
		"version", strings.TrimSpace(string(output)),
	)
	return nil
}

// CheckAllBinaries checks all required binaries and returns errors for missing ones
func CheckAllBinaries() []error {
	var errors []error

	for _, dep := range RequiredBinaries {
		if err := CheckBinary(dep); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateBinaries validates all required binaries and provides installation guidance
func ValidateBinaries() error {
	errors := CheckAllBinaries()

	if len(errors) == 0 {
		zap.S().Info("All external binary dependencies are available")
		return nil
	}

	var errorMsg strings.Builder
	errorMsg.WriteString("❌ Missing required external binaries:\n\n")

	for _, err := range errors {
		errorMsg.WriteString(fmt.Sprintf("• %s\n", err.Error()))
	}

	errorMsg.WriteString("\n💡 Installation instructions:\n")
	for _, dep := range RequiredBinaries {
		errorMsg.WriteString(fmt.Sprintf("\n%s:\n%s", dep.Name, dep.InstallHelp))
	}

	errorMsg.WriteString("\n⚠️  Please install the missing binaries and restart the bot.")
	errorMsg.WriteString("\nNote: Some features may be limited without these binaries.")

	finalError := fmt.Errorf("%s", errorMsg.String())
	zap.S().Errorw("Binary validation failed", "error", finalError)

	return finalError
}

// CheckBinaryVersion attempts to get version information for a binary
func CheckBinaryVersion(binaryName string) (string, error) {
	var cmd *exec.Cmd

	switch binaryName {
	case "tesseract":
		cmd = exec.Command("tesseract", "--version")
	case "pdftotext":
		cmd = exec.Command("pdftotext", "-v")
	default:
		return "", fmt.Errorf("unknown binary: %s", binaryName)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get version for %s: %w", binaryName, err)
	}

	version := strings.TrimSpace(string(output))

	// Extract just the version line
	lines := strings.Split(version, "\n")
	if len(lines) > 0 {
		version = strings.TrimSpace(lines[0])
	}

	return version, nil
}

// GetBinaryStatus returns the status of all required binaries
func GetBinaryStatus() map[string]BinaryStatus {
	status := make(map[string]BinaryStatus)

	for _, dep := range RequiredBinaries {
		version, err := CheckBinaryVersion(dep.Name)
		binaryStatus := BinaryStatus{
			Name:      dep.Name,
			Available: err == nil,
			Version:   version,
			Error:     err,
			Help:      dep.InstallHelp,
		}
		status[dep.Name] = binaryStatus
	}

	return status
}

// BinaryStatus represents the status of a binary dependency
type BinaryStatus struct {
	Name      string
	Available bool
	Version   string
	Error     error
	Help      string
}
