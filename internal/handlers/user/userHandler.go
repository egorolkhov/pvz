package user

import "avito/internal/manager"

type UserHandler struct {
	Mgr *manager.Manager
}

func NewUserHandler(mgr *manager.Manager) *UserHandler {
	return &UserHandler{Mgr: mgr}
}
