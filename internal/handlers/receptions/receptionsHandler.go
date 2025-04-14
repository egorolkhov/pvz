package receptions

import "avito/internal/manager"

type ReceptionsHandler struct {
	Mgr *manager.Manager
}

func NewReceptionsHandler(mgr *manager.Manager) *ReceptionsHandler {
	return &ReceptionsHandler{Mgr: mgr}
}
