package redis

import (
	"context"
	"encoding/json"
	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/internal/domain/valueObject"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type recoveryStorage struct {
	*redis.Client
	recoveryTtl time.Duration
}

type recoveryRecord struct {
	Code      string    `json:"code"`
	UserID    uuid.UUID `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func newRecoveryRecord(recovery auth.Recovery) recoveryRecord {
	return recoveryRecord{Code: recovery.Code, UserID: recovery.UserID, ExpiresAt: recovery.ExpiresAt}
}

func (r recoveryRecord) toRecovery() *auth.Recovery {
	return &auth.Recovery{Code: r.Code, UserID: r.UserID, ExpiresAt: r.ExpiresAt}
}

func NewRecoveryStorage(client *redis.Client, recoveryTtl time.Duration) auth.RecoveryStorage {
	return &recoveryStorage{
		Client:      client,
		recoveryTtl: recoveryTtl,
	}
}

func (s *recoveryStorage) CreateRecovery(ctx context.Context, userID uuid.UUID) (*auth.Recovery, error) {
	recovery := &auth.Recovery{
		Code:      uuid.New().String(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.recoveryTtl),
	}

	data, err := json.Marshal(newRecoveryRecord(*recovery))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recovery: %w", err)
	}

	key := getRecoveryKey(recovery.Code)

	if err := s.Client.Set(ctx, key, data, s.recoveryTtl).Err(); err != nil {
		return nil, err
	}

	return recovery, nil
}

func (s *recoveryStorage) GetRecovery(ctx context.Context, code string) (*auth.Recovery, error) {
	key := getRecoveryKey(code)

	result, err := s.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}
	if result == "" {
		return nil, nil
	}

	var record recoveryRecord

	err = json.Unmarshal([]byte(result), &record)
	if err != nil {
		return nil, err
	}

	return record.toRecovery(), nil
}

func (s *recoveryStorage) DeleteRecovery(ctx context.Context, code string) error {
	key := getRecoveryKey(code)

	if err := s.Client.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

func (s *recoveryStorage) AcquireCooldown(ctx context.Context, email string, cooldown time.Duration) (bool, error) {
	return s.Client.SetNX(ctx, recoveryCooldownKey(email), 1, cooldown).Result()
}

func recoveryCooldownKey(email string) string {
	return "recovery-cooldown:" + hashValue(valueObject.NormalizeEmail(email))
}

func getRecoveryKey(code string) string {
	return "recovery:" + code
}
