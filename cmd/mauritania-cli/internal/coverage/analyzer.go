package coverage

import (
	"bufio"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/validation"
)

type CoverageAnalyzer struct {
	logger *log.Logger
}

func NewCoverageAnalyzer(logger *log.Logger) *CoverageAnalyzer {
	return &CoverageAnalyzer{logger: logger}
}

func (ca *CoverageAnalyzer) AnalyzeCoverage(projectRoot string) (float64, error) {
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		ca.logger.Printf("Failed to run go test: %v, output: %s", err, string(output))
		return 0, err
	}
	lines := strings.Split(string(output), "\n")
	var coverage float64
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "coverage:" && i+1 < len(parts) {
					covStr := strings.TrimSuffix(parts[i+1], "%")
					coverage, err = strconv.ParseFloat(covStr, 64)
					if err != nil {
						return 0, err
					}
					coverage /= 100
					break
				}
			}
		}
	}
	return coverage, nil
}

func (ca *CoverageAnalyzer) GetCoverageDetails() (map[string]float64, error) {
	cmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	details := make(map[string]float64)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "\t") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				file := parts[0]
				covStr := strings.TrimSuffix(parts[2], "%")
				cov, err := strconv.ParseFloat(covStr, 64)
				if err == nil {
					details[file] = cov / 100
				}
			}
		}
	}
	return details, nil
}

func (ca *CoverageAnalyzer) UpdateFeatureCoverage(feature *validation.Feature) error {
	coverage, err := ca.AnalyzeCoverage(".")
	if err != nil {
		return err
	}
	feature.TestCoverage = coverage
	feature.LastValidated = time.Now()
	return nil
}
