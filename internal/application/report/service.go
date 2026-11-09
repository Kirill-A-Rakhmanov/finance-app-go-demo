package report

import (
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/repository"
	"time"
)

type Service struct {
	accountRepo             repository.AccountRepository
	transactionCategoryRepo repository.TransactionCategoryRepository
	counterpartyRepo        repository.CounterpartyRepository
	reportQueryService      query.ReportQueryService
	location                *time.Location
}

func NewService(
	accountRepo repository.AccountRepository,
	transactionCategoryRepo repository.TransactionCategoryRepository,
	counterpartyRepo repository.CounterpartyRepository,
	reportQueryService query.ReportQueryService,
	location *time.Location,
) *Service {
	return &Service{
		accountRepo:             accountRepo,
		transactionCategoryRepo: transactionCategoryRepo,
		counterpartyRepo:        counterpartyRepo,
		reportQueryService:      reportQueryService,
		location:                location,
	}
}
