package models

import (
	"github.com/google/uuid"
	"time"
)

const StatusClosed = "close"
const StatusInProgress = "in_progress"

type Reception struct {
	ID       uuid.UUID `json:"id,omitempty" validate:"omitempty,uuid4"`
	DateTime time.Time `json:"dateTime" validate:"required"`
	PVZID    uuid.UUID `json:"pvzId" validate:"required,uuid4"`
	Status   string    `json:"status" validate:"required,oneof=in_progress close"`
}
