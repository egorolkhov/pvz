package models

import "time"

type PvzListRequest struct {
	StartDate *time.Time `json:"startDate" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate   *time.Time `json:"endDate" time_format:"2006-01-02T15:04:05Z07:00"`
	Page      int        `json:"page" validate:"omitempty,min=1"`
	Limit     int        `json:"limit" validate:"omitempty,min=1,max=30"`
}

type ReceptionDetail struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}

type PVZDetail struct {
	PVZ        PVZ               `json:"pvz"`
	Receptions []ReceptionDetail `json:"receptions"`
}
