package repositories

import (
	"github.com/itzLilix/questboard-profile-service/internal/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)


type usersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) *usersRepository {
	return &usersRepository{db: db}
}

func (r *usersRepository) GetUserByUsername(username string) (*entities.User, error) {
	panic("unimplemented")
}
