package mapper

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/domain/model"
)

func ToUserDto(u *model.User) *dto.User {
	if u == nil {
		return nil
	}

	return &dto.User{
		ID:         u.ID,
		Email:      u.Email.ToString(),
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
	}
}
