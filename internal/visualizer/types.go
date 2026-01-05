package visualizer

// Problem represents a user-submitted problem
type Problem struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context"`
	Screenshots []string               `json:"screenshots"`
	Code        string                 `json:"code"`
	Severity    string                 `json:"severity"`
	Domain      string                 `json:"domain"`
	Tags        []string               `json:"tags"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// AnalysisResult represents AI analysis of a problem
type AnalysisResult struct {
	ProblemID        string   `json:"problem_id"`
	Patterns         []string `json:"patterns"`
	RootCauses       []string `json:"root_causes"`
	Complexity       string   `json:"complexity"`
	Impact           string   `json:"impact"`
	Domain           string   `json:"domain"`
	SuggestedActions []string `json:"suggested_actions"`
	RelatedIssues    []string `json:"related_issues"`
	Confidence       float64  `json:"confidence"`
	AnalyzedAt       string   `json:"analyzed_at"`
}
