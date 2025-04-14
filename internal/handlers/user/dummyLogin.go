package user

import (
	"avito/internal/handlers/common"
	"net/http"
)

type dummyLoginRequest struct {
	Role string `json:"role" validate:"required"`
}

func (uh *UserHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest
	err := common.DecodeAndValidate(r, &req)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	token, err := uh.Mgr.DummyLogin(r.Context(), req.Role)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	common.JsonResponse(w, http.StatusOK, token)
}
