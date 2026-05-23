package usecase

import (
	"net/mail"
	"net/url"
	"regexp"
	"unicode/utf8"
)

const (
	maxUsernameLen    = 32
	maxEmailLen       = 355
	maxDisplayNameLen = 100
	maxBioLen         = 500
	maxURLLen         = 2048
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`)

func validateEmail(email string) error {
	if email == "" || len(email) > maxEmailLen {
		return ErrInvalidData
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidData
	}
	return nil
}

func validateUsername(username string) error {
	if username == "" || len(username) > maxUsernameLen {
		return ErrInvalidData
	}
	if !usernameRegex.MatchString(username) {
		return ErrInvalidData
	}
	return nil
}

func validateDisplayName(displayName string) error {
	if displayName == "" || utf8.RuneCountInString(displayName) > maxDisplayNameLen {
		return ErrInvalidData
	}
	return nil
}

func validateURL(inputUrl string) error {
	if len(inputUrl) > maxURLLen {
		return ErrInvalidData
	}
	parsed, err := url.Parse(inputUrl)
	if err != nil {
		return ErrInvalidData
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidData
	}
	if parsed.Host == "" {
		return ErrInvalidData
	}
	return nil
}

func validateBio(bio string) error {
	if len(bio) > maxBioLen {
		return ErrInvalidData
	}
	return nil
}
