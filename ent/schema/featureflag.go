package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type FeatureFlag struct {
	ent.Schema
}

func (FeatureFlag) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.String("key").
			MaxLen(100).
			Unique(),
		field.String("description").
			MaxLen(255).
			Default(""),
		field.Bool("enabled").
			Default(false),
	}
}

func (FeatureFlag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("overrides", FeatureFlagOverride.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
