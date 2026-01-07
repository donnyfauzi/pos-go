package services

import (
	"errors"
	"pos-go/config"
	"pos-go/dto"
	user_model "pos-go/models/user_model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Sentinel errors untuk membedakan jenis kegagalan di service
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrHashPasswordFailed = errors.New("hash password failed")
	ErrCreateUserFailed   = errors.New("create user failed")
)

type AuthService interface {
	Register(input dto.RegisterDTO) (user_model.User, error)
}

type authService struct{}

func NewAuthService() AuthService {
	return &authService{}
}

func (s *authService) Register(input dto.RegisterDTO) (user_model.User, error) {
	// Validasi bisnis: cek apakah email sudah terdaftar
	var existingUser user_model.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		// Data ditemukan → email sudah dipakai
		return user_model.User{}, ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Error lain saat cek ke DB
		return user_model.User{}, ErrCreateUserFailed
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return user_model.User{}, ErrHashPasswordFailed
	}

	// Buat entity user
	user := user_model.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     input.Role,
	}

	// Simpan user ke DB
	if err := config.DB.Create(&user).Error; err != nil {
		return user_model.User{}, ErrCreateUserFailed
	}

	return user, nil
}


