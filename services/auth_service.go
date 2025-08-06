package services

import (
	"pos-go/config"
	"pos-go/dto"
	user_model "pos-go/models/user_model"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(input dto.RegisterDTO) (user_model.User, error)
}

type authService struct{}

func NewAuthService() AuthService {
	return &authService{}
}

func (s *authService) Register(input dto.RegisterDTO) (user_model.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return user_model.User{}, err
	}

	user := user_model.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     input.Role,
	}

	// Simpan user ke DB
	if err := config.DB.Create(&user).Error; err != nil {
		return user_model.User{}, err
	}

	return user, nil
}
