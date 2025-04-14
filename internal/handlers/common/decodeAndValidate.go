package common

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
)

func DecodeAndValidate(r *http.Request, dst interface{}) error {
	validate := validator.New()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("json decode error: %w", err)
	}
	defer r.Body.Close()

	if err := validate.Struct(dst); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}
