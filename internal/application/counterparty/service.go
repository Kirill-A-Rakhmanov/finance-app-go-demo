package counterparty

import "finance-app-go-demo/internal/domain/repository"

type Service struct {
	counterpartyRepo repository.CounterpartyRepository
}

func New(counterpartyRepo repository.CounterpartyRepository) *Service {
	return &Service{
		counterpartyRepo: counterpartyRepo,
	}
}
