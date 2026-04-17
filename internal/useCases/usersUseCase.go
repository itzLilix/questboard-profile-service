package usecases

type UsersUseCase interface{}

type usersUseCase struct {
	repo UsersRepository
}