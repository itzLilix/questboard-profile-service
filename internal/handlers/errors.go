package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
)

var (
	ErrUnauthorized  = errors.New("user unauthorized")
	ErrBadReq        = errors.New("bad request")
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func handleErr(c fiber.Ctx, err error) error {
	status, code, msg := resolveErr(err)
	return c.Status(status).JSON(ErrorResponse{Code: code, Message: msg})
}

func resolveErr(err error) (int, string, string) {
	switch {
	case errors.Is(err, usecase.ErrUserNotFound):
		return fiber.StatusNotFound, "USER_NOT_FOUND", "User not found"
	case errors.Is(err, usecase.ErrEmailExists):
		return fiber.StatusConflict, "EMAIL_EXISTS", "Email already registered"
	case errors.Is(err, usecase.ErrUsernameExists):
		return fiber.StatusConflict, "USERNAME_EXISTS", "Username already taken"
	case errors.Is(err, usecase.ErrWrongPassword):
		return fiber.StatusUnauthorized, "WRONG_CREDENTIALS", "Invalid email or password"
	case errors.Is(err, usecase.ErrInvalidToken):
		return fiber.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token"
	case errors.Is(err, usecase.ErrInvalidData):
		return fiber.StatusBadRequest, "INVALID_DATA", "Invalid request data"
	case errors.Is(err, usecase.ErrCannotFollowSelf):
		return fiber.StatusBadRequest, "CANNOT_FOLLOW_SELF", "Cannot follow yourself"
	case errors.Is(err, usecase.ErrFileTooLarge):
		return fiber.StatusBadRequest, "FILE_TOO_LARGE", "File exceeds size limit"
	case errors.Is(err, usecase.ErrInvalidFileType):
		return fiber.StatusBadRequest, "INVALID_FILE_TYPE", "Unsupported file type"
	case errors.Is(err, usecase.ErrInvalidCursor):
		return fiber.StatusBadRequest, "INVALID_CURSOR", "Invalid pagination cursor"
	case errors.Is(err, ErrUnauthorized):
		return fiber.StatusUnauthorized, "UNAUTHORIZED", "Authentication required"
	case errors.Is(err, ErrBadReq):
		return fiber.StatusBadRequest, "INVALID_PARAMS", "Invalid request parameters"
	}
	return fiber.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"
}

