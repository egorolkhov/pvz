package receptions

import (
	"avito/internal/handlers/common"
	"avito/internal/manager/services/pvz"
	"errors"
	"github.com/google/uuid"
	"net/http"
)

type ReceptionRequest struct {
	PvzID uuid.UUID `json:"pvzId" validate:"required"`
}

func (ph *ReceptionsHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	var req ReceptionRequest
	err := common.DecodeAndValidate(r, &req)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	reception, err := ph.Mgr.ReceptionService.Create(r.Context(), req.PvzID)
	if err != nil {
		if errors.Is(err, pvz.ErrNoPermission) {
			common.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	common.JsonResponse(w, http.StatusCreated, reception)
}
