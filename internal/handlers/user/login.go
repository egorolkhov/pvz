package user

import (
	"avito/internal/handlers/common"
	"avito/internal/manager/services/user"
	"errors"
	"net/http"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	err := common.DecodeAndValidate(r, &req)
	if err != nil {
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	token, err := uh.Mgr.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrWrongPassword) {
			common.ErrorResponse(w, http.StatusUnauthorized, err)
			return
		}
		common.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	common.JsonResponse(w, http.StatusOK, token)
}
