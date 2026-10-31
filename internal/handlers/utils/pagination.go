// Package utils provides HTTP handler utilities.
package utils

import (
	"net/http"
	"strconv"
)

// PaginationParams contains cursor-based pagination options.
type PaginationParams struct {
	Cursor string
	Limit  int
}

const (
	defaultLimit = 20
	maxLimit     = 100
)

// ParsePagination parses cursor and limit query parameters from a request.
func ParsePagination(r *http.Request) PaginationParams {
	cursor := r.URL.Query().Get("cursor")
	limit := defaultLimit

	limitQuery := r.URL.Query().Get("limit")
	if limitQuery != "" {
		if parsed, err := strconv.Atoi(limitQuery); err != nil && parsed > 0 {
			limit = parsed
		}
	}

	if limit > maxLimit {
		limit = maxLimit
	}
	return PaginationParams{
		Cursor: cursor,
		Limit:  limit,
	}
}

// PaginatedResponse contains a page of data and cursor pagination metadata.
type PaginatedResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
