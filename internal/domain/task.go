package domain

type Task struct {
	ID          *int    `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Done        *bool   `json:"done,omitempty"`
	Fk          *int    `json:"-"`
}
