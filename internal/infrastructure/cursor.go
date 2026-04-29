package infrastructure

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/itzLilix/questboard-shared/dtos"
)

type catalogCursor struct {
	Sort         dtos.UserListSort `json:"s"`
	SortOrder    dtos.SortOrder    `json:"o"`
	Rating       *float64          `json:"r,omitempty"`
	CreatedAt    *time.Time        `json:"c,omitempty"`
	FollowedAt   *time.Time        `json:"fa,omitempty"`
	ReviewsCount *int              `json:"rc,omitempty"`
	ID           string            `json:"id"`
}

func encodeCursor(c catalogCursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(s string) (*catalogCursor, error) {
	if s == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	var c catalogCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, ErrInvalidCursor
	}
	return &c, nil
}

func applyCursor(q sq.SelectBuilder, c *catalogCursor, sort dtos.UserListSort, order dtos.SortOrder) (sq.SelectBuilder, error) {
	if c == nil {
		return q, nil
	}
	if c.Sort != sort || c.SortOrder != order {
		return q, ErrInvalidCursor
	}

	op := "<"
	if order == dtos.SortAsc {
		op = ">"
	}

	switch sort {
		case dtos.SortRating:
			if c.Rating == nil {
				return q, ErrInvalidCursor
			}
			return q.Where(fmt.Sprintf("(u.rating, u.id) %s (?, ?)", op), *c.Rating, c.ID), nil
		case dtos.SortRecent:
			if c.CreatedAt == nil {
				return q, ErrInvalidCursor
			}
			return q.Where(fmt.Sprintf("(u.created_at, u.id) %s (?, ?)", op), *c.CreatedAt, c.ID), nil
		case dtos.SortFollowedAt:
			if c.FollowedAt == nil {
				return q, ErrInvalidCursor
			}
			return q.Where(fmt.Sprintf("(f.created_at, u.id) %s (?, ?)", op), *c.FollowedAt, c.ID), nil
		case dtos.SortReviewsCount:
			if c.ReviewsCount == nil {
				return q, ErrInvalidCursor
			}
			return q.Where(fmt.Sprintf("(u.reviews_count, u.id) %s (?, ?)", op), *c.ReviewsCount, c.ID), nil
		default:
			return q.Where(fmt.Sprintf("u.id %s ?", op), c.ID), nil
	}
}

func buildNextCursor(last UserCardRow, sort dtos.UserListSort, order dtos.SortOrder) (string, error) {
	c := catalogCursor{
		Sort:      sort,
		SortOrder: order,
		ID:        last.ID,
	}
	switch sort {
		case dtos.SortRating:
			v := last.Rating
			c.Rating = &v
		case dtos.SortRecent:
			v := last.CreatedAt
			c.CreatedAt = &v
		case dtos.SortFollowedAt:
			if last.FollowedAt == nil {
				return "", errors.New("build next cursor: SortFollowedAt requires non-null FollowedAt")
			}
			v := *last.FollowedAt
			c.FollowedAt = &v
		case dtos.SortReviewsCount:
			v := last.ReviewsCount
			c.ReviewsCount = &v
	}
	return encodeCursor(c)
}
