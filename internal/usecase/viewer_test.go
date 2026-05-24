package usecase

import (
	"testing"

	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/stretchr/testify/assert"
)

func TestViewer_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name string
		v    *Viewer
		want bool
	}{
		{"nil viewer", nil, false},
		{"empty user id", &Viewer{UserID: ""}, false},
		{"with user id", &Viewer{UserID: "u1"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.v.IsAuthenticated())
		})
	}
}

func TestViewer_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		v    *Viewer
		want bool
	}{
		{"nil viewer", nil, false},
		{"user role", &Viewer{UserID: "u1", Role: dtos.UserRole}, false},
		{"admin role", &Viewer{UserID: "u1", Role: dtos.AdminRole}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.v.IsAdmin())
		})
	}
}

func TestViewer_Is(t *testing.T) {
	tests := []struct {
		name   string
		v      *Viewer
		userID string
		want   bool
	}{
		{"nil viewer", nil, "u1", false},
		{"different id", &Viewer{UserID: "u1"}, "u2", false},
		{"same id", &Viewer{UserID: "u1"}, "u1", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.v.Is(tc.userID))
		})
	}
}

func TestViewer_CanActAs(t *testing.T) {
	tests := []struct {
		name    string
		v       *Viewer
		ownerID string
		want    bool
	}{
		{"unauthenticated", &Viewer{UserID: ""}, "owner", false},
		{"owner self", &Viewer{UserID: "owner", Role: dtos.UserRole}, "owner", true},
		{"non-owner user", &Viewer{UserID: "u1", Role: dtos.UserRole}, "owner", false},
		{"admin acting on someone", &Viewer{UserID: "admin", Role: dtos.AdminRole}, "owner", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.v.CanActAs(tc.ownerID))
		})
	}
}
