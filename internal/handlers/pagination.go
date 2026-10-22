// Package handlers - utilidades de paginacion
package handlers

import (
	"net/http"
	"strconv"
)

// PaginationParams structs
type PaginationParams struct {
	Cursor string
	Limit  int
}

const (
	defaultLimit = 20
	maxLimit     = 100
)

// ParsePagination parsea campos de paginacion
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

// PaginatedResponse paginado estandard con cursor
type PaginatedResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
