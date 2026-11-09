package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.String("first_name").
			MaxLen(100),
		field.String("last_name").
			MaxLen(100),
		field.String("middle_name").
			MaxLen(100).
			Optional().
			Nillable(),
		field.String("email").
			MaxLen(255).
			Unique(),
		field.String("password").
			MaxLen(255),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("accounts", Account.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("transaction_categories", TransactionCategory.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("counterparties", Counterparty.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("transactions", Transaction.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("feature_flag_overrides", FeatureFlagOverride.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
