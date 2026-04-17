package hash

import "golang.org/x/crypto/bcrypt"

type passwordHasher struct{}

func NewPasswordHasher() *passwordHasher {
	return &passwordHasher{}
}

func (h *passwordHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *passwordHasher) CompareHashAndPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
