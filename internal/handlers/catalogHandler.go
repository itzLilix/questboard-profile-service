package handlers

import (
	"fmt"
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

// UserListPage is the paginated catalog response.
type UserListPage struct {
	Items      []dtos.ProfileCardData `json:"items"`
	NextCursor string                 `json:"nextCursor,omitempty"`
}

type CatalogHandler interface {
	RegisterRoutes(router fiber.Router)
}

type catalogHandler struct {
	usecase CatalogUsecase
	log zerolog.Logger
	rbac middleware.RBACMiddleware
}

func NewCatalogHandler(usecase CatalogUsecase, log zerolog.Logger, rbac middleware.RBACMiddleware) CatalogHandler {
	return &catalogHandler{usecase: usecase, log: log, rbac: rbac}
}

func (h *catalogHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/users/", h.rbac.Optional(), h.list)
}

// @summary      List users
// @tags         catalog
// @produce      json
// @param        search      query    string   false  "Search by name or username"
// @param        format      query    string   false  "Preferred session format"    Enums(online,offline)
// @param        type        query    string   false  "Preferred session type"      Enums(oneshot,campaign)
// @param        city        query    string   false  "City"
// @param        minRating   query    number   false  "Minimum rating (0–5)"
// @param        followedBy  query    string   false  "Filter by follower (user ID or 'me')"
// @param        onlyGMs     query    boolean  false  "Only show game masters"
// @param        sort        query    string   false  "Sort field"      Enums(rating,recent,followedAt,reviews,sessions)
// @param        order       query    string   false  "Sort direction"  Enums(ASC,DESC)
// @param        cursor      query    string   false  "Pagination cursor"
// @param        limit       query    integer  false  "Page size (1–100, default 32)"
// @success      200  {object}  UserListPage
// @failure      400  {object}  ErrorResponse
// @failure      401  {object}  ErrorResponse
// @failure      500  {object}  ErrorResponse
// @router       /v1/users/ [get]
func (h *catalogHandler) list(c fiber.Ctx) error {
	var q ListUsersQuery
	viewer := viewerFromCtx(c)

	if err := c.Bind().Query(&q); err != nil {
		return handleErr(c, usecase.ErrInvalidData)
	}

	filter, err := mapQueryToFilter(&q, viewer)
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("invalid request params")
		return handleErr(c, err)
	}

	resp, err := h.usecase.ListUsers(c.Context(), viewer, *filter)
	if err != nil {
		h.log.Error().Err(err).Str("userID", viewer.UserID).Msg("list users failed")
		return handleErr(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func mapQueryToFilter(q *ListUsersQuery, viewer *usecase.Viewer) (*usecase.ListUsersFilter, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 32
	}
	if q.MinRating < 0 || q.MinRating > 5 {
		return nil, fmt.Errorf("%w: invalid filter parameter", ErrBadReq)
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
		return nil, fmt.Errorf("%w: invalid sort parameter", ErrBadReq)
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
			return "", fmt.Errorf("%w: invalid filter parameter", ErrBadReq)
	}
}

func parseSessionType(s string) (dtos.SessionType, error) {
	if s == "" {
		return "", nil
	}
	v := dtos.SessionType(s)
	switch v {
		case dtos.OneshotType, dtos.CampaignType:
			return v, nil
		default:
			return "", fmt.Errorf("%w: invalid filter parameter", ErrBadReq)
	}
}

func parseSort(s string) (dtos.UserListSort, error) {
	if s == "" {
		return "", nil
	}
	v := dtos.UserListSort(s)
	switch v {
		case dtos.SortRating, dtos.SortRecent, dtos.SortFollowedAt, dtos.SortReviewsCount, dtos.SortSessionsCount:
			return v, nil
		default:
			return "", fmt.Errorf("%w: invalid sort parameter", ErrBadReq)
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
			return "", fmt.Errorf("%w: invalid sort parameter", ErrBadReq)
	}
}