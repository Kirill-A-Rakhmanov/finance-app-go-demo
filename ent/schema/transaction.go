package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionType - enum для транзакций
type TransactionType string

const (
	TransactionTypeExpense TransactionType = "expense"
	TransactionTypeIncome  TransactionType = "income"
)

type Transaction struct {
	ent.Schema
}

func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Enum("type").
			Values("expense", "income").
			SchemaType(map[string]string{
				dialect.Postgres: "transaction_type",
			}),
		field.Float("amount").
			GoType(decimal.Decimal{}),
		field.String("name").
			MaxLen(255).
			Optional().
			Nillable(),
		field.Time("date"),
		field.Bool("is_confirmed").
			Default(false),
		field.UUID("account_id", uuid.UUID{}),
		field.UUID("transaction_category_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("counterparty_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("transfer_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).
			Ref("transactions").
			Field("account_id").
			Unique().
			Required(),
		edge.From("category", TransactionCategory.Type).
			Ref("transactions").
			Field("transaction_category_id").
			Unique(),
		edge.From("counterparty", Counterparty.Type).
			Ref("transactions").
			Field("counterparty_id").
			Unique(),
		edge.From("transfer", Transfer.Type).
			Ref("transactions").
			Field("transfer_id").
			Unique(),
	}
}

func (Transaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id", "date").
			StorageKey("idx_transactions_account_id_date").
			Annotations(entsql.DescColumns("date")),
		index.Fields("transaction_category_id").
			StorageKey("idx_transactions_transaction_category_id").
			Annotations(entsql.IndexWhere("transaction_category_id IS NOT NULL")),
		index.Fields("counterparty_id").
			StorageKey("idx_transactions_counterparty_id").
			Annotations(entsql.IndexWhere("counterparty_id IS NOT NULL")),
		index.Fields("transfer_id").
			StorageKey("idx_transactions_transfer_id").
			Annotations(entsql.IndexWhere("transfer_id IS NOT NULL")),
	}
}
