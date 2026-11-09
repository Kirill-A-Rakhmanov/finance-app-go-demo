package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/account"
	entTransaction "finance-app-go-demo/ent/transaction"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/pkg/transaction"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func toTransactionEntity(t *model.Transaction) *ent.Transaction {
	if t == nil {
		return nil
	}

	return &ent.Transaction{
		ID:                    t.ID,
		CreatedAt:             t.CreatedAt,
		UpdatedAt:             t.UpdatedAt,
		AccountID:             t.AccountID,
		TransactionCategoryID: t.TransactionCategoryID,
		TransferID:            t.TransferID,
		Type:                  entTransaction.Type(t.Type),
		Amount:                t.Amount,
		Name:                  t.Name,
		CounterpartyID:        t.CounterpartyID,
		Date:                  t.Date,
		IsConfirmed:           t.IsConfirmed,
	}
}

func toTransactionModel(t *ent.Transaction) *model.Transaction {
	if t == nil {
		return nil
	}

	return &model.Transaction{
		ID:                    t.ID,
		CreatedAt:             t.CreatedAt,
		UpdatedAt:             t.UpdatedAt,
		AccountID:             t.AccountID,
		TransactionCategoryID: t.TransactionCategoryID,
		TransferID:            t.TransferID,
		Type:                  enum.TransactionType(t.Type),
		Amount:                t.Amount,
		Name:                  t.Name,
		CounterpartyID:        t.CounterpartyID,
		Date:                  t.Date,
		IsConfirmed:           t.IsConfirmed,
	}
}

type transactionRepository struct {
	client *ent.Client
}

func NewTransactionRepository(client *ent.Client) repository.TransactionRepository {
	return &transactionRepository{client: client}
}

func (r *transactionRepository) Create(ctx context.Context, t *model.Transaction) (*model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	create := client.Transaction.Create().
		SetAccountID(t.AccountID).
		SetNillableTransactionCategoryID(t.TransactionCategoryID).
		SetNillableTransferID(t.TransferID).
		SetType(entTransaction.Type(t.Type)).
		SetAmount(t.Amount).
		SetNillableName(t.Name).
		SetNillableCounterpartyID(t.CounterpartyID).
		SetDate(t.Date).
		SetIsConfirmed(t.IsConfirmed)

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toTransactionModel(entity), nil
}

func (r *transactionRepository) BulkCreate(ctx context.Context, ts []model.Transaction) ([]model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	if len(ts) == 0 {
		return []model.Transaction{}, nil
	}

	create := client.Transaction.MapCreateBulk(ts, func(c *ent.TransactionCreate, i int) {
		c.SetAccountID(ts[i].AccountID).
			SetNillableTransactionCategoryID(ts[i].TransactionCategoryID).
			SetNillableTransferID(ts[i].TransferID).
			SetType(entTransaction.Type(ts[i].Type)).
			SetAmount(ts[i].Amount).
			SetNillableName(ts[i].Name).
			SetNillableCounterpartyID(ts[i].CounterpartyID).
			SetDate(ts[i].Date).
			SetIsConfirmed(ts[i].IsConfirmed)
	})

	entities, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Transaction, len(entities))
	for i, e := range entities {
		result[i] = *toTransactionModel(e)
	}

	return result, nil
}

func (r *transactionRepository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	client := transaction.GetClient(ctx, r.client)

	_, err := client.Transaction.Delete().
		Where(entTransaction.IDEQ(ID)).
		Exec(ctx)

	return err
}

func (r *transactionRepository) ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.Transaction.Query().
		Where(entTransaction.IDEQ(ID)).
		Exist(ctx)
}

func (r *transactionRepository) GetByID(ctx context.Context, ID uuid.UUID) (*model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Transaction.Get(ctx, ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toTransactionModel(entity), nil
}

func (r *transactionRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Transaction.Query().
		Where(entTransaction.AccountIDEQ(accountID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Transaction, len(entities))

	for i, e := range entities {
		result[i] = *toTransactionModel(e)
	}

	return result, nil
}

func (r *transactionRepository) GetByIDAndUserID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (*model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Transaction.Query().
		Where(
			entTransaction.IDEQ(ID),
			entTransaction.HasAccountWith(account.UserIDEQ(userID)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toTransactionModel(entity), nil
}

func (r *transactionRepository) GetByIDAndAccountID(ctx context.Context, ID uuid.UUID, accountID uuid.UUID) (*model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Transaction.Query().
		Where(
			entTransaction.IDEQ(ID),
			entTransaction.AccountIDEQ(accountID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toTransactionModel(entity), nil
}

func (r *transactionRepository) GetBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	client := transaction.GetClient(ctx, r.client)

	income, err := sumConfirmedAmountsByAccountID(
		ctx,
		client,
		accountID,
		entTransaction.TypeIncome,
	)
	if err != nil {
		return decimal.Zero, err
	}

	expense, err := sumConfirmedAmountsByAccountID(
		ctx,
		client,
		accountID,
		entTransaction.TypeExpense,
	)
	if err != nil {
		return decimal.Zero, err
	}

	acc, err := client.Account.Query().
		Where(account.IDEQ(accountID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return income.Sub(expense), nil
		}
		return decimal.Zero, err
	}

	return income.Sub(expense).Add(acc.Adjustment), nil
}

func (r *transactionRepository) GetBalanceByUserID(ctx context.Context, userID uuid.UUID) ([]valueObject.Balance, error) {
	client := transaction.GetClient(ctx, r.client)

	accounts, err := client.Account.Query().
		Where(
			account.UserIDEQ(userID),
			account.IsExcludedFromCountEQ(false),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return []valueObject.Balance{}, nil
	}

	accountIDs := make([]uuid.UUID, len(accounts))
	for i, acc := range accounts {
		accountIDs[i] = acc.ID
	}

	incomeByAccount, err := sumConfirmedAmountsGroupedByAccountID(
		ctx,
		client,
		accountIDs,
		entTransaction.TypeIncome,
	)
	if err != nil {
		return nil, err
	}

	expenseByAccount, err := sumConfirmedAmountsGroupedByAccountID(
		ctx,
		client,
		accountIDs,
		entTransaction.TypeExpense,
	)
	if err != nil {
		return nil, err
	}

	amountsByCurrency := make(map[valueObject.CurrencyCode]decimal.Decimal)
	currencyByCode := make(map[valueObject.CurrencyCode]valueObject.Currency)

	for _, acc := range accounts {
		currency, err := valueObject.NewCurrency(valueObject.CurrencyCode(acc.Currency))
		if err != nil {
			return nil, fmt.Errorf("failed to map account %s currency: %w", acc.ID, err)
		}

		amount := incomeByAccount[acc.ID].
			Sub(expenseByAccount[acc.ID]).
			Add(acc.Adjustment)

		amountsByCurrency[currency.Code] = amountsByCurrency[currency.Code].Add(amount)
		currencyByCode[currency.Code] = currency
	}

	balances := make([]valueObject.Balance, 0, len(amountsByCurrency))
	for code, amount := range amountsByCurrency {
		balances = append(balances, valueObject.Balance{
			Currency: currencyByCode[code],
			Amount:   amount,
		})
	}

	sort.Slice(balances, func(i, j int) bool {
		return balances[i].Currency.Code < balances[j].Currency.Code
	})

	return balances, nil
}

func sumConfirmedAmountsByAccountID(
	ctx context.Context,
	client *ent.Client,
	accountID uuid.UUID,
	_type entTransaction.Type,
) (decimal.Decimal, error) {
	var rows []struct {
		Amount decimal.Decimal `json:"amount"`
	}

	err := client.Transaction.Query().
		Where(
			entTransaction.TypeEQ(_type),
			entTransaction.AccountIDEQ(accountID),
			entTransaction.IsConfirmedEQ(true),
		).
		Aggregate(ent.As(ent.Sum(entTransaction.FieldAmount), "amount")).
		Scan(ctx, &rows)
	if err != nil {
		return decimal.Zero, err
	}
	if len(rows) == 0 {
		return decimal.Zero, nil
	}

	return rows[0].Amount, nil
}

func sumConfirmedAmountsGroupedByAccountID(
	ctx context.Context,
	client *ent.Client,
	accountIDs []uuid.UUID,
	types ...entTransaction.Type,
) (map[uuid.UUID]decimal.Decimal, error) {
	result := make(map[uuid.UUID]decimal.Decimal, len(accountIDs))
	if len(accountIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		AccountID uuid.UUID       `json:"account_id"`
		Sum       decimal.Decimal `json:"sum"`
	}

	err := client.Transaction.Query().
		Where(
			entTransaction.TypeIn(types...),
			entTransaction.AccountIDIn(accountIDs...),
			entTransaction.IsConfirmedEQ(true),
		).
		GroupBy(entTransaction.FieldAccountID).
		Aggregate(ent.As(ent.Sum(entTransaction.FieldAmount), "sum")).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.AccountID] = row.Sum
	}

	return result, nil
}

func (r *transactionRepository) GetByTransactionCategoryID(ctx context.Context, transactionCategoryID uuid.UUID) ([]model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Transaction.Query().
		Where(entTransaction.TransactionCategoryIDEQ(transactionCategoryID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Transaction, len(entities))

	for i, e := range entities {
		result[i] = *toTransactionModel(e)
	}

	return result, nil
}

func (r *transactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Transaction.Query().
		Where(entTransaction.HasAccountWith(account.UserIDEQ(userID))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Transaction, len(entities))

	for i, e := range entities {
		result[i] = *toTransactionModel(e)
	}

	return result, nil
}

func (r *transactionRepository) Update(ctx context.Context, t *model.Transaction) (*model.Transaction, error) {
	client := transaction.GetClient(ctx, r.client)

	update := client.Transaction.UpdateOneID(t.ID).
		SetAccountID(t.AccountID).
		SetType(entTransaction.Type(t.Type)).
		SetAmount(t.Amount).
		SetDate(t.Date).
		SetIsConfirmed(t.IsConfirmed)

	if t.TransactionCategoryID != nil {
		update.SetTransactionCategoryID(*t.TransactionCategoryID)
	} else {
		update.ClearTransactionCategoryID()
	}

	if t.TransferID != nil {
		update.SetTransferID(*t.TransferID)
	} else {
		update.ClearTransferID()
	}

	if t.Name != nil {
		update.SetName(*t.Name)
	} else {
		update.ClearName()
	}

	if t.CounterpartyID != nil {
		update.SetCounterpartyID(*t.CounterpartyID)
	} else {
		update.ClearCounterpartyID()
	}

	entity, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toTransactionModel(entity), nil
}
