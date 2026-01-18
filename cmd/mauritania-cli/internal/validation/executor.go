package validation

import (
	"log"
	"os/exec"
	"strings"
	"time"
)

type TestExecutor struct {
	logger *log.Logger
}

func NewTestExecutor(logger *log.Logger) *TestExecutor {
	return &TestExecutor{logger: logger}
}

func (te *TestExecutor) ExecuteTests(feature *Feature) error {
	cmd := exec.Command("go", "test", "./"+feature.Module)
	output, err := cmd.CombinedOutput()
	if err != nil {
		te.logger.Printf("Test execution failed for feature %s: %v, output: %s", feature.ID, err, string(output))
		return err
	}
	// Parse output for test results
	testCases := te.parseTestOutput(string(output))
	feature.TestCases = testCases
	allPassed := true
	for _, tc := range testCases {
		if tc.Status != "passing" {
			allPassed = false
			break
		}
	}
	if allPassed {
		feature.Status = StatusValidated
	} else {
		feature.Status = StatusFailed
	}
	feature.LastValidated = time.Now()
	return nil
}

func (te *TestExecutor) parseTestOutput(output string) []TestCase {
	var testCases []TestCase
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "=== RUN") {
			// Parse test name
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				testName := parts[2]
				tc := TestCase{
					Name:    testName,
					Type:    "unit",
					LastRun: time.Now(),
				}
				testCases = append(testCases, tc)
			}
		} else if strings.HasPrefix(line, "--- PASS") {
			if len(testCases) > 0 {
				testCases[len(testCases)-1].Status = "passing"
			}
		} else if strings.HasPrefix(line, "--- FAIL") {
			if len(testCases) > 0 {
				testCases[len(testCases)-1].Status = "failing"
			}
		}
	}
	return testCases
}
