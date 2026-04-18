package usecases

import (
	"errors"
	"fmt"
)

var (
    ErrUserNotFound    = errors.New("user not found")
    ErrWrongPassword   = errors.New("wrong password")
    ErrEmailExists     = errors.New("email already registered")
    ErrInvalidToken    = errors.New("invalid token")
	ErrUsernameExists  = errors.New("username already taken")
	ErrInvalidData	   = errors.New("invalid data")

	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrInvalidURL         = errors.New("invalid url")
	ErrInvalidBio         = errors.New("invalid bio")

	ErrCannotFollowSelf = errors.New("cannot follow yourself")
)

func wrapInvalidDataError(err error) error {
	return fmt.Errorf("%w: %s", ErrInvalidData, err)
}