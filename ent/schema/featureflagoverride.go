package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type FeatureFlagOverride struct {
	ent.Schema
}

func (FeatureFlagOverride) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.UUID("feature_flag_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Bool("enabled"),
	}
}

func (FeatureFlagOverride) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("feature_flag", FeatureFlag.Type).
			Ref("overrides").
			Field("feature_flag_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("feature_flag_overrides").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (FeatureFlagOverride) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("feature_flag_id", "user_id").
			Unique(),
		index.Fields("user_id").
			StorageKey("idx_feature_flag_overrides_user_id"),
	}
}
