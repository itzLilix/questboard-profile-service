package useCases

import "github.com/itzLilix/QuestBoard/backend/internal/repositories"

type UsersUseCase interface{}

type usersUseCase struct {
	repo repositories.UsersRepository
}