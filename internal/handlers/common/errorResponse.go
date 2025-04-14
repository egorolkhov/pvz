package common

import (
	"avito/internal/models"
	"avito/pkg/logger"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

func ErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	logger.Log.Error("internal server error", zap.Error(err))
	resp := models.Error{Message: "internal error"}
	json.NewEncoder(w).Encode(resp)
}
