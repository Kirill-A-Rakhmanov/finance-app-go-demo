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
)

// TransactionCategoryType - enum для категорий
type TransactionCategoryType string

const (
	TransactionCategoryTypeIncome  TransactionCategoryType = "income"
	TransactionCategoryTypeExpense TransactionCategoryType = "expense"
)

type TransactionCategory struct {
	ent.Schema
}

func (TransactionCategory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.String("name").
			MaxLen(100),
		field.String("slug").
			MaxLen(100).
			Optional().
			Nillable(),
		field.String("icon").
			MaxLen(100),
		field.String("color").
			MaxLen(100),
		field.Bool("is_default").
			Default(false),
		field.Enum("type").
			Values("income", "expense").
			SchemaType(map[string]string{
				dialect.Postgres: "transaction_category_type",
			}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (TransactionCategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("transaction_categories").
			Field("user_id").
			Unique().
			Required(),
		edge.To("children", TransactionCategory.Type).
			From("parent").
			Field("parent_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("transactions", Transaction.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (TransactionCategory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "type", "name").
			Unique(),
		index.Fields("parent_id").
			StorageKey("idx_transaction_categories_parent_id").
			Annotations(entsql.IndexWhere("parent_id IS NOT NULL")),
	}
}
