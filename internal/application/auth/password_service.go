package auth

type PasswordService interface {
	Hash(password string) (string, error)
	// Verify reports whether password matches hashed. An empty hashed value (no such
	// user) is checked against a dummy hash, so the call costs the same and returns false.
	Verify(password string, hashed string) bool
}
