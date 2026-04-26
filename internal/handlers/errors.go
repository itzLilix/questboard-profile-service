package handlers

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSort   error = errors.New("invalid sort parameter")
	ErrUnauthorized  error = errors.New("user unauthorized")
	ErrInvalidFilter error = errors.New("invalid filter parameter")
)

func wrapErrInvalidFilter(filter any) error {
	return fmt.Errorf("%w: %v", ErrInvalidFilter, filter)
}