package handlers

import (
	"avito/internal/handlers/pvz"
	"avito/internal/handlers/receptions"
	"avito/internal/handlers/user"
	"avito/internal/manager"
	"avito/internal/middleware"
	"avito/pkg/logger"
	"github.com/gorilla/mux"
)

func Routes(mgr *manager.Manager) *mux.Router {
	r := mux.NewRouter()
	r.Use(logger.Logger)
	r.Use(middleware.CompressMiddleware)
	r.Use(middleware.MetricsMiddleware(mgr.Metrics))

	userHandler := user.NewUserHandler(mgr)
	pvzHandler := pvz.NewPvzHandler(mgr)
	receptionHandler := receptions.NewReceptionsHandler(mgr)

	r.HandleFunc("/dummyLogin", userHandler.DummyLogin).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/register", userHandler.Register).Methods("POST")

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(mgr.TokenManager))

	protected.HandleFunc("/pvz", pvzHandler.CreatePvz).Methods("POST")
	protected.HandleFunc("/pvz", pvzHandler.PvzList).Methods("GET")

	protected.HandleFunc("/pvz/{pvzId}/close_last_reception", receptionHandler.CloseReception).Methods("POST")
	protected.HandleFunc("/pvz/{pvzId}/delete_last_product", receptionHandler.DeleteLastProduct).Methods("POST")
	protected.HandleFunc("/receptions", receptionHandler.CreateReception).Methods("POST")
	protected.HandleFunc("/products", receptionHandler.CreateProduct).Methods("POST")

	return r
}
