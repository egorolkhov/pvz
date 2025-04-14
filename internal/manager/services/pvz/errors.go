package pvz

import "errors"

var (
	ErrNoPermission  = errors.New("no permission")
	ErrForbiddenCity = errors.New("forbidden city")
)
