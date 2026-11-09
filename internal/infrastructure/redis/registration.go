package redis

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/internal/domain/valueObject"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type registrationStorage struct {
	*redis.Client
	ttl time.Duration
}

func NewRegistrationStorage(client *redis.Client, ttl time.Duration) auth.RegistrationStorage {
	return &registrationStorage{
		Client: client,
		ttl:    ttl,
	}
}

type registrationRecord struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	TokenHash    string `json:"tokenHash"`
}

func newRegistrationRecord(registration auth.PendingRegistration, tokenHash string) registrationRecord {
	return registrationRecord{
		FirstName:    registration.FirstName,
		LastName:     registration.LastName,
		Email:        registration.Email,
		PasswordHash: registration.PasswordHash,
		TokenHash:    tokenHash,
	}
}

func (r registrationRecord) toPending() auth.PendingRegistration {
	return auth.PendingRegistration{
		FirstName:    r.FirstName,
		LastName:     r.LastName,
		Email:        r.Email,
		PasswordHash: r.PasswordHash,
	}
}

func (s *registrationStorage) AcquireCooldown(ctx context.Context, email string, cooldown time.Duration) (bool, error) {
	return s.Client.SetNX(ctx, registrationCooldownKey(email), 1, cooldown).Result()
}

func (s *registrationStorage) Create(ctx context.Context, registration auth.PendingRegistration) (string, error) {
	token, err := newRegistrationToken()
	if err != nil {
		return "", err
	}

	tokenHash := hashValue(token)
	emailKey := registrationEmailKey(registration.Email)

	previous, err := s.load(ctx, emailKey)
	if err != nil {
		return "", err
	}
	if previous != nil {
		if err := s.Client.Del(ctx, registrationTokenKey(previous.TokenHash)).Err(); err != nil {
			return "", err
		}
	}

	data, err := json.Marshal(newRegistrationRecord(registration, tokenHash))
	if err != nil {
		return "", fmt.Errorf("failed to marshal registration: %w", err)
	}

	pipe := s.Client.TxPipeline()
	pipe.Set(ctx, emailKey, data, s.ttl)
	pipe.Set(ctx, registrationTokenKey(tokenHash), valueObject.NormalizeEmail(registration.Email), s.ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return "", err
	}

	return token, nil
}

func (s *registrationStorage) Get(ctx context.Context, token string) (*auth.PendingRegistration, error) {
	stored, _, err := s.find(ctx, token)
	if err != nil || stored == nil {
		return nil, err
	}

	pending := stored.toPending()

	return &pending, nil
}

func (s *registrationStorage) Delete(ctx context.Context, token string) error {
	stored, tokenHash, err := s.find(ctx, token)
	if err != nil {
		return err
	}

	pipe := s.Client.TxPipeline()
	pipe.Del(ctx, registrationTokenKey(tokenHash))
	if stored != nil {
		pipe.Del(ctx, registrationEmailKey(stored.Email))
	}

	_, err = pipe.Exec(ctx)

	return err
}

func (s *registrationStorage) find(ctx context.Context, token string) (*registrationRecord, string, error) {
	tokenHash := hashValue(token)

	email, err := s.Client.Get(ctx, registrationTokenKey(tokenHash)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, tokenHash, nil
		}

		return nil, tokenHash, err
	}

	stored, err := s.load(ctx, registrationEmailKey(email))
	if err != nil {
		return nil, tokenHash, err
	}
	if stored == nil || stored.TokenHash != tokenHash {
		return nil, tokenHash, nil
	}

	return stored, tokenHash, nil
}

func (s *registrationStorage) load(ctx context.Context, emailKey string) (*registrationRecord, error) {
	result, err := s.Client.Get(ctx, emailKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	var stored registrationRecord
	if err := json.Unmarshal([]byte(result), &stored); err != nil {
		return nil, err
	}

	return &stored, nil
}

func newRegistrationToken() (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(random), nil
}

func hashValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func registrationEmailKey(email string) string {
	return "registration:email:" + hashValue(valueObject.NormalizeEmail(email))
}

func registrationCooldownKey(email string) string {
	return "registration:cooldown:" + hashValue(valueObject.NormalizeEmail(email))
}

func registrationTokenKey(tokenHash string) string {
	return "registration:token:" + tokenHash
}
