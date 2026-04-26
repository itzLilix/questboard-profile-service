package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/rs/zerolog"
)

type ListUsersQuery struct {
	Search     string   `query:"search"`
	Systems    []string `query:"system"`
	Formats    []string `query:"format"`
	Types      []string `query:"type"`
	MinRating  float64  `query:"minRating"`
	FollowedBy string   `query:"followedBy"`
	Sort       string   `query:"sort"`
	Cursor     string   `query:"cursor"`
	Limit      int      `query:"limit"`
}

type CatalogHandler interface {
	RegisterRoutes(app *fiber.App)
}

type catalogHandler struct {
	usecase usecase.CatalogUsecase
	log zerolog.Logger
	rbac middleware.RBACMiddleware
}

func NewCatalogHandler(usecase usecase.CatalogUsecase, log zerolog.Logger, rbac middleware.RBACMiddleware) CatalogHandler {
	return &usersHandler{usecase: usecase, log: log, rbac: rbac}
}

func (h *catalogHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/users/", h.list)
}

func (h *catalogHandler) list(c fiber.Ctx) error {
	var q ListUsersQuery
	if err := c.Bind().Query(&q); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	
	viewer := viewerFromCtx(c)
	
	q.validate(viewer)
	
}

func (q *ListUsersQuery) validate(viewer *usecase.ViewerContext) error {
    if q.Limit <= 0 || q.Limit > 100 { q.Limit = 32 }
    switch q.Sort {
		case "", "rating", "recent", "followedAt":
		default: return ErrInvalidSort
    }
    if q.Sort == "followedAt" && q.FollowedBy == "" { return ErrInvalidSort }
    if q.FollowedBy == "me" && viewer.UserID == "" { return ErrUnauthorized }
    if q.MinRating < 0 || q.MinRating > 5 { return wrapErrInvalidFilter(q.MinRating) }
    return nil
}