// Package domain defines the core domain types and business rules.
package domain

import "time"

type Workspace struct {
	ID        string
	Name      string
	Slug      string
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}
