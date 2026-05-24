package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid simple", "a@b.co", false},
		{"valid with dots", "first.last@example.com", false},
		{"empty", "", true},
		{"no @", "noatsign.com", true},
		{"missing host", "user@", true},
		{"too long", strings.Repeat("a", maxEmailLen) + "@b.co", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEmail(tc.email)
			if tc.wantErr {
				assert.ErrorIs(t, err, ErrInvalidData)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"alphanum", "user123", false},
		{"with dot", "user.name", false},
		{"with dash", "user-name", false},
		{"with underscore", "user_name", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", maxUsernameLen+1), true},
		{"max length boundary", strings.Repeat("a", maxUsernameLen), false},
		{"with space", "user name", true},
		{"with special chars", "user!", true},
		{"unicode", "юзер", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUsername(tc.username)
			if tc.wantErr {
				assert.ErrorIs(t, err, ErrInvalidData)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		wantErr     bool
	}{
		{"normal", "John Doe", false},
		{"unicode counted by runes", strings.Repeat("я", maxDisplayNameLen), false},
		{"empty", "", true},
		{"too long runes", strings.Repeat("я", maxDisplayNameLen+1), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDisplayName(tc.displayName)
			if tc.wantErr {
				assert.ErrorIs(t, err, ErrInvalidData)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBio(t *testing.T) {
	tests := []struct {
		name    string
		bio     string
		wantErr bool
	}{
		{"empty allowed", "", false},
		{"normal", "Hello world", false},
		{"at limit", strings.Repeat("a", maxBioLen), false},
		{"over limit", strings.Repeat("a", maxBioLen+1), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBio(tc.bio)
			if tc.wantErr {
				assert.ErrorIs(t, err, ErrInvalidData)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https", "https://example.com", false},
		{"http", "http://example.com/path", false},
		{"too long", "https://" + strings.Repeat("a", maxURLLen) + ".com", true},
		{"no scheme", "example.com", true},
		{"ftp scheme", "ftp://example.com", true},
		{"no host", "https://", true},
		{"unparseable", "https://%zzzz", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateURL(tc.url)
			if tc.wantErr {
				assert.True(t, errors.Is(err, ErrInvalidData), "expected ErrInvalidData, got %v", err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
