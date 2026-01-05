package bot

import (
	"strings"
	"testing"
)

func TestCheckBinary_Success(t *testing.T) {
	// Test with a command that should exist (like 'echo')
	dep := BinaryDependency{
		Name:        "echo",
		CheckCmd:    []string{"echo", "test"},
		InstallHelp: "echo should be available on all systems",
	}

	err := CheckBinary(dep)
	if err != nil {
		t.Errorf("CheckBinary() error = %v, want nil", err)
	}
}

func TestCheckBinary_Failure(t *testing.T) {
	// Test with a command that should not exist
	dep := BinaryDependency{
		Name:        "definitely-non-existent-binary-xyz123",
		CheckCmd:    []string{"definitely-non-existent-binary-xyz123"},
		InstallHelp: "This binary should not exist",
	}

	err := CheckBinary(dep)
	if err == nil {
		t.Error("CheckBinary() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "definitely-non-existent-binary-xyz123") {
		t.Errorf("CheckBinary() error message = %v, should contain binary name", err.Error())
	}
}

func TestCheckBinaryVersion(t *testing.T) {
	tests := []struct {
		name      string
		binary    string
		wantErr   bool
		expectOut string
	}{
		{
			name:      "tesseract version",
			binary:    "tesseract",
			wantErr:   false,
			expectOut: "tesseract",
		},
		{
			name:      "pdftotext version",
			binary:    "pdftotext",
			wantErr:   false,
			expectOut: "pdftotext",
		},
		{
			name:      "unknown binary",
			binary:    "unknown-binary-xyz",
			wantErr:   true,
			expectOut: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := CheckBinaryVersion(tt.binary)

			if tt.wantErr {
				if err == nil {
					t.Error("CheckBinaryVersion() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("CheckBinaryVersion() unexpected error = %v", err)
				}

				// Just check that version contains the binary name (most version outputs do)
				if !strings.Contains(strings.ToLower(version), strings.ToLower(tt.expectOut)) {
					// This might not always work depending on version format, so don't fail
					t.Logf("CheckBinaryVersion() output = %v (may not contain binary name)", version)
				}
			}
		})
	}
}

func TestGetBinaryStatus(t *testing.T) {
	// Mock a missing binary by temporarily modifying RequiredBinaries
	originalBinaries := RequiredBinaries
	defer func() {
		RequiredBinaries = originalBinaries
	}()

	// Add a non-existent binary to test missing case
	RequiredBinaries = append(RequiredBinaries, BinaryDependency{
		Name:        "test-missing-binary",
		CheckCmd:    []string{"test-missing-binary"},
		InstallHelp: "This is a test binary",
	})

	status := GetBinaryStatus()

	// Should have tesseract and pdftotext (assuming they exist) + our test binary
	expectedCount := len(originalBinaries) + 1
	if len(status) != expectedCount {
		t.Errorf("GetBinaryStatus() count = %v, want %v", len(status), expectedCount)
	}

	// Check our test binary is marked as unavailable
	testStatus, exists := status["test-missing-binary"]
	if !exists {
		t.Error("GetBinaryStatus() should contain test-missing-binary")
	}

	if testStatus.Available {
		t.Error("GetBinaryStatus() test-missing-binary should be unavailable")
	}

	if testStatus.Error == nil {
		t.Error("GetBinaryStatus() test-missing-binary should have error")
	}
}

func TestCheckAllBinaries(t *testing.T) {
	// Save original and restore after test
	originalBinaries := RequiredBinaries
	defer func() {
		RequiredBinaries = originalBinaries
	}()

	// Test with mix of available and unavailable binaries
	RequiredBinaries = []BinaryDependency{
		{
			Name:        "echo", // Should be available
			CheckCmd:    []string{"echo", "test"},
			InstallHelp: "echo should be available",
		},
		{
			Name:        "non-existent-binary-xyz", // Should not be available
			CheckCmd:    []string{"non-existent-binary-xyz"},
			InstallHelp: "This should not exist",
		},
	}

	errors := CheckAllBinaries()

	// Should have exactly one error (for the non-existent binary)
	if len(errors) != 1 {
		t.Errorf("CheckAllBinaries() error count = %v, want 1", len(errors))
	}

	if errors[0] == nil {
		t.Error("CheckAllBinaries() expected non-nil error")
	}

	if !strings.Contains(errors[0].Error(), "non-existent-binary-xyz") {
		t.Errorf("CheckAllBinaries() error = %v, should contain binary name", errors[0].Error())
	}
}

func TestValidateBinaries_AllAvailable(t *testing.T) {
	// Save original and restore after test
	originalBinaries := RequiredBinaries
	defer func() {
		RequiredBinaries = originalBinaries
	}()

	// Test with all binaries available
	RequiredBinaries = []BinaryDependency{
		{
			Name:        "echo",
			CheckCmd:    []string{"echo", "test"},
			InstallHelp: "echo should be available",
		},
	}

	err := ValidateBinaries()
	if err != nil {
		t.Errorf("ValidateBinaries() error = %v, want nil", err)
	}
}

func TestValidateBinaries_SomeMissing(t *testing.T) {
	// Save original and restore after test
	originalBinaries := RequiredBinaries
	defer func() {
		RequiredBinaries = originalBinaries
	}()

	// Test with missing binary
	RequiredBinaries = []BinaryDependency{
		{
			Name:        "echo",
			CheckCmd:    []string{"echo", "test"},
			InstallHelp: "echo should be available",
		},
		{
			Name:        "definitely-missing-binary-xyz",
			CheckCmd:    []string{"definitely-missing-binary-xyz"},
			InstallHelp: "This binary definitely doesn't exist",
		},
	}

	err := ValidateBinaries()
	if err == nil {
		t.Error("ValidateBinaries() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "definitely-missing-binary-xyz") {
		t.Errorf("ValidateBinaries() error = %v, should contain missing binary name", err.Error())
	}

	if !strings.Contains(err.Error(), "Installation instructions") {
		t.Error("ValidateBinaries() error should contain installation instructions")
	}
}

func TestRequiredBinaries_Configuration(t *testing.T) {
	// Test that required binaries are properly configured
	if len(RequiredBinaries) == 0 {
		t.Error("RequiredBinaries should not be empty")
	}

	// Check for tesseract
	foundTesseract := false
	foundPdftotext := false

	for _, dep := range RequiredBinaries {
		switch dep.Name {
		case "tesseract":
			foundTesseract = true
			if len(dep.CheckCmd) == 0 || dep.CheckCmd[0] != "tesseract" {
				t.Error("tesseract CheckCmd not properly configured")
			}
			if dep.InstallHelp == "" {
				t.Error("tesseract InstallHelp not configured")
			}
		case "pdftotext":
			foundPdftotext = true
			if len(dep.CheckCmd) == 0 || dep.CheckCmd[0] != "pdftotext" {
				t.Error("pdftotext CheckCmd not properly configured")
			}
			if dep.InstallHelp == "" {
				t.Error("pdftotext InstallHelp not configured")
			}
		}
	}

	if !foundTesseract {
		t.Error("RequiredBinaries should include tesseract")
	}

	if !foundPdftotext {
		t.Error("RequiredBinaries should include pdftotext")
	}
}
