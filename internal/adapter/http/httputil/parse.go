package httputil

import (
	"time"

	"github.com/google/uuid"
)

func ParseOptionalUUID(raw *string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}

	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func ParseUUIDs(raw []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(raw))

	for _, r := range raw {
		if r == "" {
			continue
		}

		id, err := uuid.Parse(r)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func MillisToTime(ms *int64) *time.Time {
	if ms == nil {
		return nil
	}

	t := time.UnixMilli(*ms)
	return &t
}
