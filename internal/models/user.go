package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Role      string    `json:"role" db:"role"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser creates a new user with generated ID
func NewUser(email, name, role string) *User {
	now := time.Now().UTC()
	return &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Role:      role,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UserCreateRequest represents a request to create a user
type UserCreateRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Role  string `json:"role" validate:"required,oneof=admin user viewer"`
}

// UserUpdateRequest represents a request to update a user
type UserUpdateRequest struct {
	Email *string `json:"email" validate:"omitempty,email"`
	Name  *string `json:"name" validate:"omitempty,min=2,max=100"`
	Role  *string `json:"role" validate:"omitempty,oneof=admin user viewer"`
	Active *bool  `json:"active" validate:"omitempty"`
}

// UserResponse represents a user API response
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users      []*UserResponse `json:"users"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}