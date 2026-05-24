package usecase

import (
	"errors"
	"testing"

	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-shared/cursor"
	"github.com/stretchr/testify/assert"
)

func TestWrapInvalidDataError(t *testing.T) {
	inner := errors.New("bad shape")
	err := wrapInvalidDataError(inner)
	assert.ErrorIs(t, err, ErrInvalidData)
	assert.Contains(t, err.Error(), "bad shape")
}

func TestMapRepoErr(t *testing.T) {
	tests := []struct {
		name    string
		in      error
		want    error
		wantNil bool
	}{
		{"nil", nil, nil, true},
		{"user not found", infrastructure.ErrUserNotFound, ErrUserNotFound, false},
		{"duplicate username", infrastructure.ErrDuplicateUsername, ErrUsernameExists, false},
		{"duplicate email", infrastructure.ErrDuplicateEmail, ErrEmailExists, false},
		{"cannot follow self", infrastructure.ErrCannotFollowSelf, ErrCannotFollowSelf, false},
		{"no new data", infrastructure.ErrNoNewData, ErrInvalidData, false},
		{"invalid cursor", cursor.ErrInvalidCursor, ErrInvalidCursor, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mapRepoErr("op", tc.in)
			if tc.wantNil {
				assert.NoError(t, got)
				return
			}
			assert.ErrorIs(t, got, tc.want)
		})
	}
}

func TestMapRepoErr_UnknownWrapsInternal(t *testing.T) {
	got := mapRepoErr("some op", errors.New("boom"))
	assert.ErrorIs(t, got, ErrInternal)
	assert.Contains(t, got.Error(), "some op")
}
