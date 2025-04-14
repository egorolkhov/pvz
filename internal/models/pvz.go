package models

import (
	"github.com/google/uuid"
	"time"
)

type PVZ struct {
	ID               uuid.UUID `json:"id,omitempty"`
	RegistrationDate time.Time `json:"registrationDate,omitempty"`
	City             string    `json:"city" validate:"required"`
}
