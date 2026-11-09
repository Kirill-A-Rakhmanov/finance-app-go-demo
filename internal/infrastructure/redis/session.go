package redis

import (
	"context"
	"encoding/json"
	"finance-app-go-demo/internal/application/auth"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type sessionStorage struct {
	*redis.Client
	sessionTtl time.Duration
}

type sessionRecord struct {
	ID        string    `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func newSessionRecord(session auth.Session) sessionRecord {
	return sessionRecord{ID: session.ID, UserID: session.UserID, ExpiresAt: session.ExpiresAt}
}

func (r sessionRecord) toSession() *auth.Session {
	return &auth.Session{ID: r.ID, UserID: r.UserID, ExpiresAt: r.ExpiresAt}
}

func NewSessionStorage(client *redis.Client, sessionTtl time.Duration) auth.SessionStorage {
	return &sessionStorage{
		Client:     client,
		sessionTtl: sessionTtl,
	}
}

func (s *sessionStorage) CreateSession(ctx context.Context, userID uuid.UUID) (*auth.Session, error) {
	session := &auth.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.sessionTtl),
	}

	data, err := json.Marshal(newSessionRecord(*session))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	pipe := s.Client.TxPipeline()
	pipe.Set(ctx, getKey(session.ID), data, s.sessionTtl)
	pipe.SAdd(ctx, getUserKey(userID), session.ID)
	pipe.Expire(ctx, getUserKey(userID), s.sessionTtl)

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionStorage) DeleteSession(ctx context.Context, ID string) error {
	session, err := s.GetSession(ctx, ID)
	if err != nil {
		return err
	}

	pipe := s.Client.TxPipeline()
	pipe.Del(ctx, getKey(ID))
	if session != nil {
		pipe.SRem(ctx, getUserKey(session.UserID), ID)
	}

	_, err = pipe.Exec(ctx)

	return err
}

var deleteUserSessionsScript = redis.NewScript(`
local ids = redis.call('SMEMBERS', KEYS[1])
for _, id in ipairs(ids) do
	if id ~= ARGV[2] then
		redis.call('DEL', ARGV[1] .. id)
		redis.call('SREM', KEYS[1], id)
	end
end
return #ids
`)

func (s *sessionStorage) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, exceptSessionID string) error {
	return deleteUserSessionsScript.Run(ctx, s.Client, []string{getUserKey(userID)}, getKey(""), exceptSessionID).Err()
}

func (s *sessionStorage) ShortenSession(ctx context.Context, ID string, ttl time.Duration) error {
	return s.Client.ExpireLT(ctx, getKey(ID), ttl).Err()
}

func (s *sessionStorage) GetSession(ctx context.Context, ID string) (*auth.Session, error) {
	key := getKey(ID)

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

	var record sessionRecord

	err = json.Unmarshal([]byte(result), &record)
	if err != nil {
		return nil, err
	}

	return record.toSession(), nil
}

func getKey(sessionID string) string {
	return "session:" + sessionID
}

func getUserKey(userID uuid.UUID) string {
	return "user_sessions:" + userID.String()
}
