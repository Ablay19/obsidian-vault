# Branch Commit Review Data Model

## Overview

The data model for the branch commit review feature consists of core entities that track repository state, commit information, and review findings. All entities include validation rules derived from functional requirements.

## Entities

### Branch Entity

**Purpose**: Represents a Git branch with its current state and metadata

**Fields**:
- `id`: String (UUID, primary key)
- `name`: String (branch name, required, unique per repository)
- `repository_url`: String (Git repository URL, required)
- `status`: Enum (active, stale, merged, deleted, required)
- `last_commit_hash`: String (SHA-1 hash of latest commit)
- `last_commit_date`: DateTime (timestamp of last commit)
- `last_commit_author`: String (author of last commit)
- `ahead_count`: Integer (commits ahead of base branch, default 0)
- `behind_count`: Integer (commits behind base branch, default 0)
- `has_merge_conflicts`: Boolean (true if conflicts exist, default false)
- `created_at`: DateTime (auto-generated)
- `updated_at`: DateTime (auto-updated)

**Validation Rules**:
- Name must be valid Git branch name
- Repository URL must be accessible Git URL
- Status transitions: active ↔ stale, active → merged, any → deleted
- Ahead/behind counts must be non-negative

**Relationships**:
- One-to-many with Commit (commits in this branch)
- Many-to-one with Repository (containing repository)

### Commit Entity

**Purpose**: Represents a Git commit with metadata and analysis results

**Fields**:
- `id`: String (UUID, primary key)
- `hash`: String (SHA-1 hash, required, unique)
- `branch_id`: String (foreign key to Branch)
- `author_name`: String (commit author name, required)
- `author_email`: String (commit author email, required)
- `commit_date`: DateTime (commit timestamp, required)
- `message`: String (commit message, required)
- `parent_hashes`: Array<String> (parent commit hashes)
- `changed_files`: Array<String> (list of modified files)
- `insertions`: Integer (lines added, default 0)
- `deletions`: Integer (lines removed, default 0)
- `quality_score`: Integer (0-100, calculated)
- `has_security_issues`: Boolean (true if security problems found)
- `is_large_commit`: Boolean (true if >1000 lines changed)
- `sensitive_data_found`: Boolean (true if secrets detected)
- `created_at`: DateTime (auto-generated)

**Validation Rules**:
- Hash must be valid SHA-1
- Message must be non-empty, <500 characters
- Quality score must be 0-100
- Changed files array limited to 1000 entries
- Insertions/deletions must be non-negative

**Relationships**:
- Many-to-one with Branch (containing branch)
- One-to-many with Issue (issues found in this commit)

### Review Entity

**Purpose**: Represents a review session with findings and recommendations

**Fields**:
- `id`: String (UUID, primary key)
- `target_type`: Enum (branch, commit_range, repository, required)
- `target_identifier`: String (branch name, commit range, or repo URL)
- `review_type`: Enum (individual, comprehensive, security, required)
- `status`: Enum (pending, in_progress, completed, failed, required)
- `started_at`: DateTime (review start time)
- `completed_at`: DateTime (review completion time)
- `total_commits`: Integer (commits analyzed, default 0)
- `issues_found`: Integer (total issues identified, default 0)
- `critical_issues`: Integer (high-severity issues, default 0)
- `recommendations`: Array<String> (suggested improvements)
- `report_url`: String (link to generated report)
- `reviewer`: String (user who performed review)
- `created_at`: DateTime (auto-generated)
- `updated_at`: DateTime (auto-updated)

**Validation Rules**:
- Target identifier required and validated based on type
- Status transitions: pending → in_progress → completed/failed
- Issues counts must be non-negative
- Recommendations limited to 50 items

**Relationships**:
- One-to-many with Issue (issues from this review)
- Many-to-one with Branch/Repository (review target)

### Issue Entity

**Purpose**: Represents a specific problem or finding from analysis

**Fields**:
- `id`: String (UUID, primary key)
- `review_id`: String (foreign key to Review)
- `commit_hash`: String (foreign key to Commit)
- `issue_type`: Enum (security, quality, performance, documentation, required)
- `severity`: Enum (critical, high, medium, low, info, required)
- `category`: String (specific issue category, e.g., "SQL injection")
- `description`: String (detailed description, required)
- `file_path`: String (affected file path)
- `line_number`: Integer (line number in file)
- `code_snippet`: String (relevant code excerpt)
- `remediation`: String (suggested fix)
- `tool_name`: String (analysis tool that found issue)
- `confidence`: Float (0.0-1.0, tool confidence level)
- `created_at`: DateTime (auto-generated)

**Validation Rules**:
- Description required, <1000 characters
- Confidence must be 0.0-1.0
- Line number must be positive if provided
- Code snippet limited to 500 characters

**Relationships**:
- Many-to-one with Review (containing review)
- Many-to-one with Commit (related commit)

## Data Flow

1. **Repository Scan**: Extract Branch and Commit entities
2. **Analysis**: Create Issue entities linked to Commits
3. **Review Session**: Create Review entity with aggregated Issues
4. **Reporting**: Generate reports from Review and related entities

## Constraints

- All commits must belong to exactly one branch
- Reviews can span multiple commits but belong to one target
- Issues are always tied to specific commits within reviews
- Entity relationships maintain referential integrity
- Historical data preserved for audit trails