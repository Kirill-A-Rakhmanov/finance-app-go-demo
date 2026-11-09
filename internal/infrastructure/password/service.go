package password

import (
	"finance-app-go-demo/internal/application/auth"

	"golang.org/x/crypto/bcrypt"
)

type passwordService struct {
	cost      int
	dummyHash []byte
}

func NewPasswordService(cost int) auth.PasswordService {
	dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), cost)

	return &passwordService{cost: cost, dummyHash: dummyHash}
}

func (s *passwordService) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (s *passwordService) Verify(password, hashed string) bool {
	if hashed == "" {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}
