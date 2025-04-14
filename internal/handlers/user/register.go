package user

import (
	"avito/internal/handlers/common"
	"net/http"
)

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=employee moderator"`
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	err := common.DecodeAndValidate(r, &req)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user, err := uh.Mgr.Register(r.Context(), req.Email, req.Role, req.Password)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	common.JsonResponse(w, http.StatusCreated, user)
}
