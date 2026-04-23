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
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validateUsername(username string) error {
	if username == "" || len(username) > maxUsernameLen {
		return ErrInvalidUsername
	}
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

func validateDisplayName(displayName string) error {
	if displayName == "" || utf8.RuneCountInString(displayName) > maxDisplayNameLen {
		return ErrInvalidDisplayName
	}
	return nil
}

func validateURL(inputUrl string) error {
	if len(inputUrl) > maxURLLen {
		return ErrInvalidURL
	}
	parsed, err := url.Parse(inputUrl)
	if err != nil {
		return ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}
	if parsed.Host == "" {
		return ErrInvalidURL
	}
	return nil
}

func validateBio(bio string) error {
	if len(bio) > maxBioLen {
		return ErrInvalidBio
	}
	return nil
}
