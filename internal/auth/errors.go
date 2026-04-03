package auth

import "errors"

var (
    ErrUserNotFound    = errors.New("user not found")
    ErrWrongPassword   = errors.New("wrong password")
    ErrEmailExists      = errors.New("email already registered")
    ErrInvalidToken    = errors.New("invalid token")
	ErrUsernameExists  = errors.New("username already taken")
)