package auth

import "time"

// todo: decide where limits and TTLs live; some are env (token and registration TTLs, API rate limit), the ones below are constants
const (
	MaxFailedAttempts    = 5
	FailedAttemptsWindow = 15 * time.Minute
	RecoveryCooldown     = 2 * time.Minute
	RegistrationCooldown = 10 * time.Minute
	RefreshGracePeriod   = 30 * time.Second
)
