package pvz

import "avito/internal/manager"

type PvzHandler struct {
	Mgr *manager.Manager
}

func NewPvzHandler(mgr *manager.Manager) *PvzHandler {
	return &PvzHandler{Mgr: mgr}
}
