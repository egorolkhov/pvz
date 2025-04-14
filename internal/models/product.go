package models

import (
	"github.com/google/uuid"
	"time"
)

type Product struct {
	ID          uuid.UUID `json:"id,omitempty" validate:"omitempty,uuid4"`
	DateTime    time.Time `json:"dateTime,omitempty"`
	Type        string    `json:"type" validate:"required,oneof=электроника одежда обувь"`
	ReceptionID uuid.UUID `json:"receptionId" validate:"required,uuid4"`
}
