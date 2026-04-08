package repositories

import (
	"github.com/itzLilix/QuestBoard/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository interface {
	GetUserByUsername(username string) (*models.User, error)
}

type usersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) UsersRepository {
	return &usersRepository{db: db}
}

func (r *usersRepository) GetUserByUsername(username string) (*models.User, error) {
	panic("unimplemented")
}
