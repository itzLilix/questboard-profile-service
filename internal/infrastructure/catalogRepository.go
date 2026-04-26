package infrastructure

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCatalogFilter struct {
    Search       string
    Systems      []string
    Formats      []dtos.SessionFormat
    Types        []dtos.SessionType
    MinRating    float64
    FollowedByID string
    Sort         dtos.SortOrder
    Cursor       string
    Limit        int
}

type catalogRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}