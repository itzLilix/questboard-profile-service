package usecase

type CatalogUsecase interface {
}

type catalogUsecase struct {
	repo CatalogRepository
}