package usecase

import (
	"errors"
	"fmt"

	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/cursor"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrWrongPassword  = errors.New("wrong password")
	ErrEmailExists    = errors.New("email already registered")
	ErrUsernameExists = errors.New("username already taken")
	ErrInvalidToken   = errors.New("invalid token")
	ErrInvalidData    = errors.New("invalid data")
	ErrInternal       = errors.New("internal server error")

	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	ErrInvalidCursor    = errors.New("invalid cursor")
	ErrFileTooLarge     = errors.New("file too large")
	ErrInvalidFileType  = errors.New("invalid file type")
)

func wrapInvalidDataError(err error) error {
	return fmt.Errorf("%w: %s", ErrInvalidData, err)
}

func mapRepoErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, infrastructure.ErrUserNotFound) {
		return ErrUserNotFound
	}
	if errors.Is(err, infrastructure.ErrDuplicateUsername) {
		return ErrUsernameExists
	}
	if errors.Is(err, infrastructure.ErrDuplicateEmail) {
		return ErrEmailExists
	}
	if errors.Is(err, infrastructure.ErrCannotFollowSelf) {
		return ErrCannotFollowSelf
	}
	if errors.Is(err, infrastructure.ErrNoNewData) {
		return ErrInvalidData
	}
	if errors.Is(err, cursor.ErrInvalidCursor) {
		return ErrInvalidCursor
	}
	return fmt.Errorf("%s: %w", op, ErrInternal)
}
