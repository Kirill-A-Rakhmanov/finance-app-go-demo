package model

import (
	"finance-app-go-demo/internal/domain/enum"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	AccountID             uuid.UUID
	TransactionCategoryID *uuid.UUID
	CounterpartyID        *uuid.UUID
	TransferID            *uuid.UUID
	Type                  enum.TransactionType
	Amount                decimal.Decimal
	Name                  *string
	Date                  time.Time
	IsConfirmed           bool
}

func NewTransaction(
	accountID uuid.UUID,
	transactionCategoryID *uuid.UUID,
	counterpartyID *uuid.UUID,
	transferID *uuid.UUID,
	type_ enum.TransactionType,
	amount decimal.Decimal,
	name *string,
	date time.Time,
	isConfirmed bool,
) *Transaction {
	return &Transaction{
		AccountID:             accountID,
		TransactionCategoryID: transactionCategoryID,
		CounterpartyID:        counterpartyID,
		TransferID:            transferID,
		Type:                  type_,
		Amount:                amount,
		Name:                  name,
		Date:                  date,
		IsConfirmed:           isConfirmed,
	}
}

func NewExpenseTransaction(
	accountID uuid.UUID,
	transactionCategoryID *uuid.UUID,
	counterpartyID *uuid.UUID,
	amount decimal.Decimal,
	name *string,
	date time.Time,
	isConfirmed bool,
) *Transaction {
	return NewTransaction(
		accountID,
		transactionCategoryID,
		counterpartyID,
		nil,
		enum.TransactionExpense,
		amount,
		name,
		date,
		isConfirmed,
	)
}

func NewIncomeTransaction(
	accountID uuid.UUID,
	transactionCategoryID *uuid.UUID,
	counterpartyID *uuid.UUID,
	amount decimal.Decimal,
	name *string,
	date time.Time,
	isConfirmed bool,
) *Transaction {
	return NewTransaction(
		accountID,
		transactionCategoryID,
		counterpartyID,
		nil,
		enum.TransactionIncome,
		amount,
		name,
		date,
		isConfirmed,
	)
}

func NewTransferInTransaction(
	accountID uuid.UUID,
	amount decimal.Decimal,
	name string,
	date time.Time,
	isConfirmed bool,
	transferID uuid.UUID,
) *Transaction {
	return NewTransaction(
		accountID,
		nil,
		nil,
		&transferID,
		enum.TransactionIncome,
		amount,
		&name,
		date,
		isConfirmed,
	)
}

func NewTransferOutTransaction(
	accountID uuid.UUID,
	amount decimal.Decimal,
	name string,
	date time.Time,
	isConfirmed bool,
	transferID uuid.UUID,
) *Transaction {
	return NewTransaction(
		accountID,
		nil,
		nil,
		&transferID,
		enum.TransactionExpense,
		amount,
		&name,
		date,
		isConfirmed,
	)
}

func BalanceEffect(type_ enum.TransactionType, amount decimal.Decimal, isConfirmed bool) decimal.Decimal {
	if !isConfirmed {
		return decimal.Zero
	}

	if type_ == enum.TransactionExpense {
		return amount.Neg()
	}

	return amount
}

func (t *Transaction) BalanceEffect() decimal.Decimal {
	return BalanceEffect(t.Type, t.Amount, t.IsConfirmed)
}
