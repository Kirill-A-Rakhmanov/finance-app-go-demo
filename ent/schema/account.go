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

type Account struct {
	ent.Schema
}

func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.String("name").
			MaxLen(255),
		field.String("icon").
			MaxLen(255),
		field.String("currency").
			MaxLen(3),
		field.Bool("is_negative_balance_enabled").
			Default(false),
		field.Bool("is_hidden").
			Default(false),
		field.Bool("is_excluded_from_count").
			Default(false),
		field.Float("adjustment").
			GoType(decimal.Decimal{}),
		field.UUID("user_id", uuid.UUID{}),
	}
}

func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("accounts").
			Field("user_id").
			Unique().
			Required(),
		edge.To("transactions", Transaction.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("from_transfers", Transfer.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("to_transfers", Transfer.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").
			StorageKey("idx_accounts_user_id"),
	}
}
