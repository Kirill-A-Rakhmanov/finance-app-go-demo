package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/account"
	"finance-app-go-demo/ent/predicate"
	entTransaction "finance-app-go-demo/ent/transaction"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/pkg/transaction"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type transactionQueryService struct {
	client *ent.Client
}

func NewTransactionQueryService(client *ent.Client) query.TransactionQueryService {
	return &transactionQueryService{client: client}
}

func (qs *transactionQueryService) GetTransactions(ctx context.Context, criteria query.GetTransactionCriteria) (*query.GetTransactionsResult, error) {
	client := transaction.GetClient(ctx, qs.client)

	predicates := buildTransactionPredicates(criteria)
	order := buildTransactionOrder(criteria)

	total, err := client.Transaction.Query().
		Where(predicates...).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &query.GetTransactionsResult{
			Total:        0,
			Order:        make([]uuid.UUID, 0),
			Transactions: make([]model.Transaction, 0),
		}, nil
	}

	entities, err := client.Transaction.Query().
		Where(predicates...).
		Order(order).
		Limit(criteria.Limit).
		Offset(criteria.Offset).
		All(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(entities))
	transactions := make([]model.Transaction, len(entities))

	for i, e := range entities {
		ids[i] = e.ID
		transactions[i] = *toTransactionModel(e)
	}

	return &query.GetTransactionsResult{
		Total:        total,
		Order:        ids,
		Transactions: transactions,
	}, nil
}

func buildTransactionPredicates(criteria query.GetTransactionCriteria) []predicate.Transaction {
	predicates := make([]predicate.Transaction, 0)

	predicates = append(predicates, entTransaction.HasAccountWith(account.UserIDEQ(criteria.UserID)))

	if criteria.PeriodStart != nil {
		predicates = append(predicates, entTransaction.DateGTE(*criteria.PeriodStart))
	}

	if criteria.PeriodEnd != nil {
		predicates = append(predicates, entTransaction.DateLTE(*criteria.PeriodEnd))
	}

	if criteria.AccountID != nil {
		predicates = append(predicates, entTransaction.AccountIDEQ(*criteria.AccountID))
	}

	if len(criteria.TransactionCategoryIDs) > 0 {
		predicates = append(predicates, entTransaction.TransactionCategoryIDIn(criteria.TransactionCategoryIDs...))
	}

	return predicates
}

func buildTransactionOrder(criteria query.GetTransactionCriteria) entTransaction.OrderOption {
	if criteria.SortBy == nil || criteria.SortDirection == nil {
		return entTransaction.ByCreatedAt(sql.OrderDesc())
	}

	switch *criteria.SortBy {
	case query.TransactionSortByDate:
		if *criteria.SortDirection == query.Ascending {
			return entTransaction.ByDate(sql.OrderAsc())
		}

		return entTransaction.ByDate(sql.OrderDesc())

	case query.TransactionSortByAmount:
		if *criteria.SortDirection == query.Ascending {
			return entTransaction.ByAmount(sql.OrderAsc())
		}

		return entTransaction.ByAmount(sql.OrderDesc())

	default:
		return entTransaction.ByCreatedAt(sql.OrderDesc())
	}
}
