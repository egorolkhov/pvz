package pvz

import (
	"avito/internal/handlers/common"
	"avito/internal/manager/services/pvz"
	"avito/internal/models"
	"errors"
	"net/http"
)

func (ph *PvzHandler) CreatePvz(w http.ResponseWriter, r *http.Request) {
	var req models.PVZ
	err := common.DecodeAndValidate(r, &req)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	pvzCreated, err := ph.Mgr.PvzService.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, pvz.ErrNoPermission) {
			common.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	common.JsonResponse(w, http.StatusCreated, pvzCreated)
}
