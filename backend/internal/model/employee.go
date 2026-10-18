package model

import "time"

// Allowed employee statuses.
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// Employee is the core domain entity.
type Employee struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Department  string    `json:"department"`
	Designation string    `json:"designation"`
	JoiningDate string    `json:"joining_date"` // YYYY-MM-DD
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EmployeeInput is the payload for create and update requests.
type EmployeeInput struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
	JoiningDate string `json:"joining_date"`
	Status      string `json:"status"`
}

// ListFilter holds optional list query parameters.
type ListFilter struct {
	Search     string // matches name or email
	Department string
	Status     string
}
