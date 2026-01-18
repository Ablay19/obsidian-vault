# Data Model

## Core Entities

### Feature

Represents a functionality or component in the codebase that needs validation.

```go
type Feature struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    Module       string    `json:"module"`
    Status       string    `json:"status"` // "validated", "failed", "pending"
    TestCoverage float64   `json:"test_coverage"`
    LastValidated time.Time `json:"last_validated"`
    TestCases    []TestCase `json:"test_cases"`
}
```

**Validation Rules**:
- Name is required and unique within module
- TestCoverage must be between 0.0 and 1.0
- Status must be one of: "validated", "failed", "pending"
- Module must reference existing module directory

**State Transitions**:
```
pending → validated (when all tests pass)
pending → failed (when any test fails)
failed → validated (when issues are fixed)
```

### Document

Represents documentation files and their completeness status.

```go
type Document struct {
    ID              string    `json:"id"`
    Type            string    `json:"type"` // "README", "API", "GUIDE", "CODE_COMMENT"
    Location        string    `json:"location"`
    LastUpdated     time.Time `json:"last_updated"`
    CompletenessScore float64 `json:"completeness_score"`
    CoverageFeature  string    `json:"coverage_feature"`
    MaintainedBy    string    `json:"maintained_by"`
}
```

**Validation Rules**:
- Type must be one of: "README", "API", "GUIDE", "CODE_COMMENT"
- Location must be valid file path
- CompletenessScore between 0.0 and 1.0
- MaintainedBy must reference team member

### Directory

Represents project directories and their organization status.

```go
type Directory struct {
    ID          string    `json:"id"`
    Path        string    `json:"path"`
    Purpose     string    `json:"purpose"`
    FileCount   int       `json:"file_count"`
    LastCleaned time.Time `json:"last_cleaned"`
    Module      string    `json:"module"`
}
```

**Validation Rules**:
- Path must be valid directory path
- FileCount must be non-negative
- Purpose must describe directory's role
- Module must reference parent module

## Test Coverage Entity

### TestCase

Represents individual test cases for validation.

```go
type TestCase struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    FeatureID   string `json:"feature_id"`
    Type        string `json:"type"` // "unit", "integration", "e2e"
    Status      string `json:"status"` // "passing", "failing", "skipped"
    LastRun     time.Time `json:"last_run"`
    Duration    time.Duration `json:"duration"`
}
```

## Relationships

- **Feature → TestCase**: One-to-many (Feature has multiple TestCases)
- **Feature → Module**: Many-to-one (Feature belongs to one Module)
- **Document → Feature**: Many-to-one (Document covers one Feature)
- **Document → TeamMember**: Many-to-one (Document maintained by one TeamMember)
- **Directory → Module**: Many-to-one (Directory part of one Module)
- **Directory → Files**: One-to-many (Directory contains multiple Files)

## Data Flow

### Validation Process
1. **Discovery**: Scan codebase to identify Features
2. **Test Execution**: Run tests for each Feature
3. **Coverage Analysis**: Calculate TestCoverage for Features
4. **Status Update**: Set Feature.Status based on test results
5. **Timestamp**: Update LastValidated timestamp

### Documentation Process
1. **Scanning**: Identify all documentation files
2. **Analysis**: Calculate completeness for each Document
3. **Gap Identification**: Find missing documentation types
4. **Generation**: Create missing Document entries
5. **Maintenance**: Update LastUpdated and CompletenessScore

### Cleanup Process
1. **Directory Analysis**: Scan Directory entities
2. **File Counting**: Update FileCount for each Directory
3. **Cleanup Execution**: Remove unnecessary files
4. **Status Update**: Set LastCleaned timestamp
5. **Validation**: Ensure directory structure follows conventions

## Storage Requirements

### File-based Storage
- **Features**: `features.json` - List of all features and their validation status
- **Documents**: `documents.json` - Documentation inventory and completeness
- **Directories**: `directories.json` - Directory structure analysis
- **Coverage Reports**: `coverage.out`, `coverage.html` - Test coverage data

### Temporary Storage
- **Test Results**: Stored in memory during validation, persisted as summary
- **Cleanup Cache**: Temporary list of files identified for removal
- **Documentation Drafts**: Working copies during generation

## Data Integrity

### Constraints
- Feature.TestCoverage cannot exceed 1.0 (100%)
- Document.CompletenessScore cannot exceed 1.0 (100%)
- Directory.FileCount must match actual file count on disk
- All timestamps must be UTC for consistency

### Validation
- Schema validation for all JSON files
- Path validation for file locations
- Cross-reference validation between entities
- Historical data integrity checks

This data model provides a comprehensive foundation for tracking functionality validation, documentation status, and directory cleanup operations while maintaining data integrity and supporting the feature's success criteria.