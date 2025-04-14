package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID `json:"id,omitempty"`
	CreatedAt    time.Time `json:"-"`
	Email        string    `json:"email" validate:"required,email"`
	Role         string    `json:"role" validate:"required,oneof=employee moderator"`
	PasswordHash string    `json:"-"`
}
