package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transfer struct {
	ent.Schema
}

func (Transfer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Float("from_amount").
			GoType(decimal.Decimal{}),
		field.Float("to_amount").
			GoType(decimal.Decimal{}),
		field.Time("date"),
		field.UUID("from_account_id", uuid.UUID{}),
		field.UUID("to_account_id", uuid.UUID{}),
	}
}

func (Transfer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("from_account", Account.Type).
			Ref("from_transfers").
			Field("from_account_id").
			Unique().
			Required(),
		edge.From("to_account", Account.Type).
			Ref("to_transfers").
			Field("to_account_id").
			Unique().
			Required(),
		edge.To("transactions", Transaction.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Transfer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("from_account_id").
			StorageKey("idx_transfers_from_account_id"),
		index.Fields("to_account_id").
			StorageKey("idx_transfers_to_account_id"),
	}
}
