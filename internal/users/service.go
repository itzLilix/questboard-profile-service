package users

type Service interface{}

type service struct {
	repo Repository
}