package repositories

import "errors"

var (
    ErrDuplicateEmail    = errors.New("duplicate email")
    ErrDuplicateUsername = errors.New("duplicate username")
    ErrUserNotFound      = errors.New("user not found")
)