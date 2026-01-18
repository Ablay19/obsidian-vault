# Branch Commit Review Quickstart

## Overview

Get started with comprehensive Git branch and commit analysis in minutes. This guide covers installation, basic usage, and integration examples.

## Prerequisites

- Go 1.21+ installed
- Git repository access (local or remote)
- Basic command-line knowledge

## Installation

### From Source

```bash
git clone https://github.com/your-org/branch-review.git
cd branch-review
go build -o branch-review ./cmd
```

### Using Go Install

```bash
go install github.com/your-org/branch-review/cmd@latest
```

## Basic Usage

### Review a Local Repository

```bash
# Navigate to your Git repository
cd /path/to/your/repo

# Review current branch
branch-review analyze --branch main

# Review all branches
branch-review analyze --all-branches

# Generate security-focused report
branch-review analyze --security --output security-report.html
```

### CLI Commands

```bash
# Show help
branch-review --help

# List available branches
branch-review branches

# Analyze specific commit range
branch-review analyze --range abc123..def456

# Generate detailed report
branch-review report --format pdf --output review.pdf
```

## Configuration

### Basic Configuration

Create a `.branch-review.yaml` file in your repository:

```yaml
repository:
  url: "https://github.com/your-org/repo"
  default_branch: "main"

analysis:
  enable_security: true
  enable_quality: true
  max_commit_size: 1000

reporting:
  default_format: "html"
  include_recommendations: true
```

### Environment Variables

```bash
# GitHub token for private repositories
export GITHUB_TOKEN=your_token_here

# Custom analysis tools path
export BRANCH_REVIEW_TOOLS_PATH=/path/to/tools

# Output directory
export BRANCH_REVIEW_OUTPUT_DIR=./reports
```

## Integration Examples

### GitHub Actions

Add to your `.github/workflows/review.yml`:

```yaml
name: Branch Review
on:
  pull_request:
    branches: [ main, develop ]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Run Branch Review
      run: |
        curl -L https://github.com/your-org/branch-review/releases/download/v1.0.0/branch-review-linux-amd64 -o branch-review
        chmod +x branch-review
        ./branch-review analyze --branch ${{ github.head_ref }} --output pr-report.html
    - name: Upload Report
      uses: actions/upload-artifact@v3
      with:
        name: review-report
        path: pr-report.html
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    stages {
        stage('Branch Review') {
            steps {
                sh '''
                    wget https://github.com/your-org/branch-review/releases/download/v1.0.0/branch-review-linux-amd64
                    chmod +x branch-review
                    ./branch-review analyze --all-branches --output ${WORKSPACE}/review-report.html
                '''
                publishHTML(target: [
                    allowMissing: false,
                    alwaysLinkToLastBuild: true,
                    keepAll: true,
                    reportDir: '.',
                    reportFiles: 'review-report.html',
                    reportName: 'Branch Review Report'
                ])
            }
        }
    }
}
```

### GitLab CI

```yaml
branch_review:
  stage: test
  script:
    - wget https://github.com/your-org/branch-review/releases/download/v1.0.0/branch-review-linux-amd64 -O branch-review
    - chmod +x branch-review
    - ./branch-review analyze --branch $CI_COMMIT_REF_NAME --output review-report.html
  artifacts:
    paths:
      - review-report.html
    expire_in: 1 week
  only:
    - merge_requests
```

## Advanced Usage

### Custom Analysis Rules

Create a `.branch-review/rules.yaml` file:

```yaml
rules:
  - name: "Large Commits"
    type: "commit"
    condition: "changes > 500"
    severity: "medium"
    message: "Commit contains more than 500 changes"

  - name: "Security Keywords"
    type: "content"
    pattern: "(?i)(password|secret|key|token)"
    severity: "high"
    message: "Potential sensitive data exposure"
```

### API Integration

```bash
# Start web server
branch-review server --port 8080

# Trigger review via API
curl -X POST http://localhost:8080/api/v1/reviews \
  -H "Content-Type: application/json" \
  -d '{
    "targetType": "branch",
    "targetIdentifier": "feature/new-feature",
    "reviewType": "comprehensive"
  }'
```

### Webhook Setup

Configure webhooks for automated reviews:

```bash
# GitHub webhook
branch-review webhook github --secret your-webhook-secret

# Generic webhook
branch-review webhook generic --url http://your-ci-server/webhook
```

## Output Formats

### HTML Report
Interactive web-based report with:
- Branch overview dashboard
- Commit timeline visualization
- Issue severity charts
- Detailed recommendations

### PDF Report
Printable comprehensive report with:
- Executive summary
- Detailed findings
- Code snippets
- Remediation steps

### JSON/SARIF
Machine-readable formats for:
- CI/CD integration
- External tool consumption
- Automated processing

## Troubleshooting

### Common Issues

**Permission Denied**
```bash
# Ensure Git repository access
git config --global credential.helper store
# Or use SSH keys for private repos
ssh-keygen -t rsa -b 4096 -C "your-email@example.com"
```

**Analysis Timeout**
```yaml
# Increase timeout in config
analysis:
  timeout: 300  # 5 minutes
```

**Large Repository**
```yaml
# Optimize for large repos
analysis:
  batch_size: 100
  parallel_workers: 4
```

### Getting Help

- Documentation: https://docs.branch-review.example.com
- Issues: https://github.com/your-org/branch-review/issues
- Discussions: https://github.com/your-org/branch-review/discussions

## Next Steps

1. Run your first analysis on a test repository
2. Configure custom rules for your team's standards
3. Integrate with your CI/CD pipeline
4. Set up automated reviews for pull requests
5. Explore advanced features like security scanning

For more detailed documentation, visit the [full user guide](https://docs.branch-review.example.com/guide).