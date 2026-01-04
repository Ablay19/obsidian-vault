package util

import (
	"fmt"
	"os/exec"

	"go.uber.org/zap"
)

// CheckExternalBinaries verifies that required external binaries are available in the system's PATH.
func CheckExternalBinaries() error {
	requiredBinaries := []string{"tesseract", "pdftotext"}
	var missingBinaries []string

	for _, binary := range requiredBinaries {
		_, err := exec.LookPath(binary)
		if err != nil {
			missingBinaries = append(missingBinaries, binary)
		}
	}

	if len(missingBinaries) > 0 {
		return fmt.Errorf("missing required external binaries: %v. Please install them and ensure they are in your system's PATH", missingBinaries)
	}

	zap.S().Info("All required external binaries found.")
	return nil
}
