package receptions

import (
	"avito/internal/handlers/common"
	"avito/internal/manager/services/pvz"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
)

func (ph *ReceptionsHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pvzIDStr, ok := vars["pvzId"]
	if !ok {
		common.ErrorResponse(w, http.StatusBadRequest, errors.New("wrongParams"))
		return
	}

	pvzID, err := uuid.Parse(pvzIDStr)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = ph.Mgr.ReceptionService.DeleteLastProduct(r.Context(), pvzID)
	if err != nil {
		if errors.Is(err, pvz.ErrNoPermission) {
			common.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
