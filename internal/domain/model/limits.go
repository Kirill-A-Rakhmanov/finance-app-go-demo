package model

import "github.com/shopspring/decimal"

const (
	AccountNameMaxLen = 50
	AccountIconMaxLen = 255

	TransactionNameMaxLen = 255

	CounterpartyNameMaxLen = 100

	FeatureFlagKeyMaxLen         = 100
	FeatureFlagDescriptionMaxLen = 255

	TransactionCategoryNameMaxLen  = 50
	TransactionCategoryIconMaxLen  = 100
	TransactionCategoryColorMaxLen = 100

	UserFirstNameMaxLen = 100
	UserLastNameMaxLen  = 100
	UserEmailMaxLen     = 255
)

const AmountMaxScale = 4

var AmountMax = decimal.New(1, 12)
