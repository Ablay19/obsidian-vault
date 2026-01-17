package pagination

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// PaginationConfig holds pagination configuration
type PaginationConfig struct {
	DefaultPageSize int    `json:"default_page_size" yaml:"default_page_size"`
	MaxPageSize     int    `json:"max_page_size" yaml:"max_page_size"`
	DefaultSortBy   string `json:"default_sort_by" yaml:"default_sort_by"`
	DefaultSortDir  string `json:"default_sort_dir" yaml:"default_sort_dir"`
}

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	SortBy   string `json:"sort_by" form:"sort_by"`
	SortDir  string `json:"sort_dir" form:"sort_dir"`
	Cursor   string `json:"cursor" form:"cursor"`
}

// PaginationResult holds pagination result metadata
type PaginationResult struct {
	CurrentPage int         `json:"current_page"`
	PageSize    int         `json:"page_size"`
	TotalPages  int         `json:"total_pages"`
	TotalItems  int64       `json:"total_items"`
	HasNext     bool        `json:"has_next"`
	HasPrev     bool        `json:"has_prev"`
	NextCursor  string      `json:"next_cursor,omitempty"`
	PrevCursor  string      `json:"prev_cursor,omitempty"`
	Data        interface{} `json:"data"`
}

// CursorPaginationResult holds cursor-based pagination result
type CursorPaginationResult struct {
	Data       interface{} `json:"data"`
	NextCursor string      `json:"next_cursor,omitempty"`
	HasNext    bool        `json:"has_next"`
	Count      int         `json:"count"`
}

// NewPaginationConfig creates default pagination configuration
func NewPaginationConfig() *PaginationConfig {
	return &PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
		DefaultSortBy:   "created_at",
		DefaultSortDir:  "desc",
	}
}

// ParsePaginationParams parses pagination parameters from HTTP request
func ParsePaginationParams(r *http.Request, config *PaginationConfig) (*PaginationParams, error) {
	params := &PaginationParams{
		Page:     config.DefaultPageSize,
		PageSize: config.DefaultPageSize,
		SortBy:   config.DefaultSortBy,
		SortDir:  config.DefaultSortDir,
	}

	// Parse query parameters
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			if pageSize > config.MaxPageSize {
				params.PageSize = config.MaxPageSize
			} else {
				params.PageSize = pageSize
			}
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}

	if sortDir := r.URL.Query().Get("sort_dir"); sortDir != "" {
		if sortDir == "asc" || sortDir == "desc" {
			params.SortDir = sortDir
		}
	}

	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		params.Cursor = cursor
	}

	return params, nil
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(params *PaginationParams, config *PaginationConfig) error {
	if params.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	if params.PageSize < 1 {
		return fmt.Errorf("page_size must be greater than 0")
	}
	if params.PageSize > config.MaxPageSize {
		return fmt.Errorf("page_size cannot exceed %d", config.MaxPageSize)
	}
	if params.SortDir != "asc" && params.SortDir != "desc" {
		return fmt.Errorf("sort_dir must be 'asc' or 'desc'")
	}
	return nil
}

// CalculateOffset calculates the database offset for pagination
func CalculateOffset(params *PaginationParams) int {
	return (params.Page - 1) * params.PageSize
}

// CreatePaginationResult creates a pagination result from query results
func CreatePaginationResult(data interface{}, totalItems int64, params *PaginationParams) *PaginationResult {
	totalPages := int(math.Ceil(float64(totalItems) / float64(params.PageSize)))

	result := &PaginationResult{
		CurrentPage: params.Page,
		PageSize:    params.PageSize,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasNext:     params.Page < totalPages,
		HasPrev:     params.Page > 1,
		Data:        data,
	}

	return result
}

// CreateCursorPaginationResult creates a cursor-based pagination result
func CreateCursorPaginationResult(data interface{}, nextCursor string, hasNext bool) *CursorPaginationResult {
	return &CursorPaginationResult{
		Data:       data,
		NextCursor: nextCursor,
		HasNext:    hasNext,
		Count:      len(data.([]interface{})), // This is a simplification
	}
}

// EncodeCursor encodes a cursor value (typically ID + timestamp)
func EncodeCursor(id interface{}, timestamp interface{}) string {
	return fmt.Sprintf("%v:%v", id, timestamp)
}

// DecodeCursor decodes a cursor value
func DecodeCursor(cursor string) (id string, timestamp string, err error) {
	parts := strings.Split(cursor, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid cursor format")
	}
	return parts[0], parts[1], nil
}

// BuildOrderByClause builds SQL ORDER BY clause from pagination parameters
func BuildOrderByClause(params *PaginationParams, allowedSortFields []string) string {
	// Validate sort field
	isAllowed := false
	for _, field := range allowedSortFields {
		if params.SortBy == field {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		params.SortBy = "created_at" // fallback
	}

	return fmt.Sprintf("%s %s", params.SortBy, strings.ToUpper(params.SortDir))
}

// BuildLimitClause builds SQL LIMIT clause
func BuildLimitClause(params *PaginationParams) string {
	return fmt.Sprintf("LIMIT %d", params.PageSize)
}

// BuildOffsetClause builds SQL OFFSET clause
func BuildOffsetClause(params *PaginationParams) string {
	offset := CalculateOffset(params)
	return fmt.Sprintf("OFFSET %d", offset)
}

// BuildCursorWhereClause builds SQL WHERE clause for cursor-based pagination
func BuildCursorWhereClause(cursor string, sortField string, sortDir string) (string, error) {
	if cursor == "" {
		return "", nil
	}

	id, timestamp, err := DecodeCursor(cursor)
	if err != nil {
		return "", err
	}

	var operator string
	if sortDir == "desc" {
		operator = "<"
	} else {
		operator = ">"
	}

	if sortField == "created_at" {
		return fmt.Sprintf("(%s %s '%s' OR (%s = '%s' AND id %s '%s'))",
			sortField, operator, timestamp,
			sortField, timestamp,
			operator, id), nil
	}

	return fmt.Sprintf("%s %s '%s'", sortField, operator, id), nil
}

// SetPaginationHeaders sets pagination-related HTTP headers
func SetPaginationHeaders(w http.ResponseWriter, result *PaginationResult) {
	w.Header().Set("X-Total-Count", fmt.Sprintf("%d", result.TotalItems))
	w.Header().Set("X-Total-Pages", fmt.Sprintf("%d", result.TotalPages))
	w.Header().Set("X-Current-Page", fmt.Sprintf("%d", result.CurrentPage))
	w.Header().Set("X-Page-Size", fmt.Sprintf("%d", result.PageSize))
	w.Header().Set("X-Has-Next", fmt.Sprintf("%t", result.HasNext))
	w.Header().Set("X-Has-Prev", fmt.Sprintf("%t", result.HasPrev))
}

// SetCursorPaginationHeaders sets cursor pagination HTTP headers
func SetCursorPaginationHeaders(w http.ResponseWriter, result *CursorPaginationResult) {
	w.Header().Set("X-Has-Next", fmt.Sprintf("%t", result.HasNext))
	w.Header().Set("X-Count", fmt.Sprintf("%d", result.Count))
	if result.NextCursor != "" {
		w.Header().Set("X-Next-Cursor", result.NextCursor)
	}
}

// ValidateSortField validates if a sort field is allowed
func ValidateSortField(sortBy string, allowedFields []string) bool {
	for _, field := range allowedFields {
		if sortBy == field {
			return true
		}
	}
	return false
}

// Common allowed sort fields for different entities
var (
	UserSortFields = []string{
		"created_at", "updated_at", "email", "name",
	}

	ProcessingFileSortFields = []string{
		"created_at", "updated_at", "file_name", "status", "file_size",
	}

	ConversationSortFields = []string{
		"created_at", "updated_at", "title",
	}

	MessageSortFields = []string{
		"created_at", "updated_at", "role",
	}

	NotificationSortFields = []string{
		"created_at", "updated_at", "title", "status", "notification_type",
	}

	WebhookDeliverySortFields = []string{
		"created_at", "updated_at", "status", "response_status",
	}

	UsageSortFields = []string{
		"date", "requests", "tokens", "cost",
	}
)

// PaginationMiddleware creates a middleware for automatic pagination handling
func PaginationMiddleware(config *PaginationConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse pagination parameters
			params, err := ParsePaginationParams(r, config)
			if err != nil {
				http.Error(w, "Invalid pagination parameters", http.StatusBadRequest)
				return
			}

			// Validate parameters
			if err := ValidatePaginationParams(params, config); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Store pagination params in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "pagination_params", params)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetPaginationParamsFromContext extracts pagination parameters from request context
func GetPaginationParamsFromContext(ctx context.Context) *PaginationParams {
	if params, ok := ctx.Value("pagination_params").(*PaginationParams); ok {
		return params
	}
	return nil
}

// PaginatedQuery represents a paginated database query
type PaginatedQuery struct {
	SelectClause string
	FromClause   string
	WhereClause  string
	OrderBy      string
	Limit        int
	Offset       int
}

// BuildSQL builds the complete SQL query
func (pq *PaginatedQuery) BuildSQL() string {
	query := fmt.Sprintf("SELECT %s FROM %s", pq.SelectClause, pq.FromClause)

	if pq.WhereClause != "" {
		query += fmt.Sprintf(" WHERE %s", pq.WhereClause)
	}

	if pq.OrderBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", pq.OrderBy)
	}

	if pq.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", pq.Limit)
	}

	if pq.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", pq.Offset)
	}

	return query
}

// BuildCountSQL builds the count query for total items
func (pq *PaginatedQuery) BuildCountSQL() string {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.FromClause)

	if pq.WhereClause != "" {
		query += fmt.Sprintf(" WHERE %s", pq.WhereClause)
	}

	return query
}
