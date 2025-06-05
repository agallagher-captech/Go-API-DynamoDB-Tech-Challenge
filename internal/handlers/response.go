package handlers

import "github.com/google/uuid"

// userResponse represents the output model for a user.
type userResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}
