package repositories

import (
	"github.com/itzLilix/questboard-shared/models"
	"github.com/jackc/pgx/v5/pgxpool"
)


type usersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) *usersRepository {
	return &usersRepository{db: db}
}

func (r *usersRepository) GetUserByUsername(username string) (*models.User, error) {
	panic("unimplemented")
}
