package domain

type Task struct {
	Id          *int    `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Done        *bool   `json:"done,omitempty"`
	Fk          *int    `json:"-"`
}
