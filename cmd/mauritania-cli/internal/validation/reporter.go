package validation

import (
	"encoding/json"
	"os"
	"text/template"
)

type ValidationReporter struct {
}

func NewValidationReporter() *ValidationReporter {
	return &ValidationReporter{}
}

func (vr *ValidationReporter) GenerateReport(features []Feature, outputPath string) error {
	report := struct {
		TotalFeatures     int
		ValidatedFeatures int
		FailedFeatures    int
		PendingFeatures   int
		AverageCoverage   float64
		Features          []Feature
	}{
		TotalFeatures: len(features),
		Features:      features,
	}
	totalCoverage := 0.0
	for _, f := range features {
		totalCoverage += f.TestCoverage
		switch f.Status {
		case StatusValidated:
			report.ValidatedFeatures++
		case StatusFailed:
			report.FailedFeatures++
		case StatusPending:
			report.PendingFeatures++
		}
	}
	if len(features) > 0 {
		report.AverageCoverage = totalCoverage / float64(len(features))
	}
	tmpl := `
Validation Report
=================

Total Features: {{.TotalFeatures}}
Validated: {{.ValidatedFeatures}}
Failed: {{.FailedFeatures}}
Pending: {{.PendingFeatures}}
Average Coverage: {{printf "%.2f" .AverageCoverage}}%

Features:
{{range .Features}}
- {{.Name}} ({{.Module}}): {{.Status}} - Coverage: {{printf "%.2f" .TestCoverage}}
{{end}}
`
	t := template.Must(template.New("report").Parse(tmpl))
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return t.Execute(file, report)
}

func (vr *ValidationReporter) GenerateJSONReport(features []Feature, outputPath string) error {
	return json.NewEncoder(os.Stdout).Encode(features)
}
