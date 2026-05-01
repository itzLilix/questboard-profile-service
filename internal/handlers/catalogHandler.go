package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/rs/zerolog"
)

type ListUsersQuery struct {
	Search     string   `query:"search"`
	Format     string 	`query:"format"`
	Type       string 	`query:"type"`
	City 	   string   `query:"city"`
	MinRating  float64  `query:"minRating"`
	FollowedBy string   `query:"followedBy"`
	OnlyGMs	   bool		`query:"onlyGMs"`
	Sort       string   `query:"sort"`
	SortOrder  string 	`query:"order"`
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
	return &catalogHandler{usecase: usecase, log: log, rbac: rbac}
}

func (h *catalogHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/users/", h.rbac.Optional(), h.list)
}

func (h *catalogHandler) list(c fiber.Ctx) error {
	var q ListUsersQuery
	viewer := viewerFromCtx(c)

	if err := c.Bind().Query(&q); err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("unable bind request params")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	filter, err := mapQueryToFilter(&q, viewer)
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("invalid request params")
		if errors.Is(err, ErrUnauthorized) {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.SendStatus(fiber.StatusBadRequest)
	}

	resp, err := h.usecase.ListUsers(viewer, *filter)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, usecase.ErrInvalidCursor) {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("error listing users")
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func mapQueryToFilter(q *ListUsersQuery, viewer *usecase.ViewerContext) (*usecase.ListUsersFilter, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 32
	}
	if q.MinRating < 0 || q.MinRating > 5 {
		return nil, wrapErrInvalidFilter(q.MinRating)
	}

	format, err := parseSessionFormat(q.Format)
	if err != nil {
		return nil, err
	}
	sessionType, err := parseSessionType(q.Type)
	if err != nil {
		return nil, err
	}
	sort, err := parseSort(q.Sort)
	if err != nil {
		return nil, err
	}
	sortOrder, err := parseSortOrder(q.SortOrder)
	if err != nil {
		return nil, err
	}

	if q.FollowedBy == "me" && viewer.UserID == "" {
		return nil, ErrUnauthorized
	}
	
	if sort == dtos.SortFollowedAt && q.FollowedBy == "" {
		return nil, ErrInvalidSort
	}

	return &usecase.ListUsersFilter{
		Search:     q.Search,
		Format:     format,
		Type:       sessionType,
		City:       q.City,
		MinRating:  q.MinRating,
		FollowedBy: q.FollowedBy,
		OnlyGMs: 	q.OnlyGMs,
		Sort:       sort,
		SortOrder:  sortOrder,
		Cursor:     q.Cursor,
		Limit:      uint64(q.Limit),
	}, nil
}

func parseSessionFormat(s string) (dtos.SessionFormat, error) {
	if s == "" {
		return "", nil
	}
	v := dtos.SessionFormat(s)
	switch v {
		case dtos.Online, dtos.Offline:
			return v, nil
		default:
			return "", ErrInvalidFilter
	}
}

func parseSessionType(s string) (dtos.SessionType, error) {
	if s == "" {
		return "", nil
	}
	v := dtos.SessionType(s)
	switch v {
		case dtos.Oneshot, dtos.Campaign:
			return v, nil
		default:
			return "", ErrInvalidFilter
	}
}

func parseSort(s string) (dtos.UserListSort, error) {
	if s == "" {
		return "", nil
	}
	v := dtos.UserListSort(s)
	switch v {
		case dtos.SortRating, dtos.SortRecent, dtos.SortFollowedAt, dtos.SortSessionsCount:
			return v, nil
		default:
			return "", ErrInvalidSort
	}
}

func parseSortOrder(s string) (dtos.SortOrder, error) {
	if s == "" {
		return dtos.SortDesc, nil
	}

	v := dtos.SortOrder(strings.ToUpper(s))
	switch v {
		case dtos.SortAsc, dtos.SortDesc:
			return v, nil
		default:
			return "", ErrInvalidSort
	}
}