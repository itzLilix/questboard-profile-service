package users

import (
	"github.com/itzLilix/QuestBoard/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetUserByUsername(username string) (*models.User, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetUserByUsername(username string) (*models.User, error) 
