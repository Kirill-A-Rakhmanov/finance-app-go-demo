package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Counterparty struct {
	ent.Schema
}

func (Counterparty) Fields() []ent.Field {
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
		field.UUID("user_id", uuid.UUID{}),
	}
}

func (Counterparty) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("counterparties").
			Field("user_id").
			Unique().
			Required(),
		edge.To("transactions", Transaction.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (Counterparty) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "name").
			Unique(),
	}
}
