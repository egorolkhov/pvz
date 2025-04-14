package reception

import "errors"

var (
	ErrReception            = errors.New("reception error")
	ErrOpenReception        = errors.New("you have open reception")
	ErrNoOpenReception      = errors.New("you don't have open reception")
	ErrNoProductToDelete    = errors.New("no product to delete")
	ErrForbiddenProductType = errors.New("forbidden product type")
)
