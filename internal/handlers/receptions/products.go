package receptions

import (
	"avito/internal/handlers/common"
	"avito/internal/manager/services/pvz"
	"errors"
	"github.com/google/uuid"
	"net/http"
)

type ProductsRequest struct {
	Type  string    `json:"type" validate:"required,oneof=электроника одежда обувь"`
	PvzID uuid.UUID `json:"pvzId" validate:"required"`
}

func (ph *ReceptionsHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req ProductsRequest
	err := common.DecodeAndValidate(r, &req)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	product, err := ph.Mgr.ReceptionService.AddProduct(r.Context(), req.Type, req.PvzID)
	if err != nil {
		if errors.Is(err, pvz.ErrNoPermission) {
			common.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	common.JsonResponse(w, http.StatusCreated, product)
}
