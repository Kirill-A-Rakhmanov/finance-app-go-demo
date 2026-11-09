package model

import (
	"finance-app-go-demo/internal/domain/valueObject"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	Email      valueObject.Email
	Password   string
	FirstName  string
	LastName   string
	MiddleName *string
}

func NewUser(
	firstName string,
	lastName string,
	email valueObject.Email,
	password string,
	middleName *string,
) *User {
	return &User{
		FirstName:  firstName,
		LastName:   lastName,
		Email:      email,
		Password:   password,
		MiddleName: middleName,
	}
}
