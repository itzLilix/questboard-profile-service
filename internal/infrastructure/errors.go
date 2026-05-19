package infrastructure

import (
	"errors"
)

var (
    ErrDuplicateEmail         = errors.New("duplicate email")
    ErrDuplicateUsername      = errors.New("duplicate username")
    ErrUserNotFound           = errors.New("user not found")
    ErrRefreshTokenNotFound   = errors.New("refresh token not found")
    ErrCannotFollowSelf       = errors.New("cannot follow yourself")
	ErrNoNewData			  = errors.New("no new data in request")
)