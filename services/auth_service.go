package services

import (
	"errors"
	"pos-go/config"
	"pos-go/dto"
	user_model "pos-go/models/user_model"
	"pos-go/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Sentinel errors untuk membedakan jenis kegagalan di service
var (
	ErrEmailAlreadyExists = errors.New("Email sudah terdaftar")
	ErrHashPasswordFailed = errors.New("Gagal memproses password")
	ErrCreateUserFailed   = errors.New("Gagal membuat user baru")
	ErrInvalidCredentials = errors.New("Email atau password salah")
	ErrInvalidOldPassword = errors.New("Password lama tidak cocok")
	ErrGetUsersFailed     = errors.New("Gagal mengambil daftar user")
	ErrUserNotFound       = errors.New("User tidak ditemukan")
	ErrDeleteUserFailed   = errors.New("Gagal menghapus user")
)

type AuthService interface {
	Register(input dto.RegisterDTO) (user_model.User, error)
	Login(input dto.LoginDTO) (user_model.User, string, error)
	ChangePassword(userID string, input dto.ChangePasswordDTO) error
	GetCurrentUser(userID string) (user_model.User, error)
	GetAllUsers() ([]user_model.User, error)
	DeleteUser(userID string) error
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

func (s *authService) Login(input dto.LoginDTO) (user_model.User, string, error) {
	// 1. Cari user berdasarkan email
	var user user_model.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User tidak ditemukan
			return user_model.User{}, "", ErrInvalidCredentials
		}
		// Error lain dari database
		return user_model.User{}, "", ErrCreateUserFailed
	}

	// 2. Bandingkan password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		// Password tidak cocok
		return user_model.User{}, "", ErrInvalidCredentials
	}

	// 3. Generate JWT token
	token, err := utils.GenerateToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return user_model.User{}, "", ErrCreateUserFailed
	}

	// 4. Return user dan token
	return user, token, nil
}

func (s *authService) ChangePassword(userID string, input dto.ChangePasswordDTO) error {
	// 1. Cari user berdasarkan ID
	var user user_model.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCreateUserFailed
		}
		return ErrCreateUserFailed
	}

	// 2. Verify old password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword))
	if err != nil {
		return ErrInvalidOldPassword
	}

	// 3. Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return ErrHashPasswordFailed
	}

	// 4. Update password
	user.Password = string(hashedPassword)
	if err := config.DB.Save(&user).Error; err != nil {
		return ErrCreateUserFailed
	}

	return nil
}

func (s *authService) GetCurrentUser(userID string) (user_model.User, error) {
	var user user_model.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user_model.User{}, ErrCreateUserFailed
		}
		return user_model.User{}, ErrCreateUserFailed
	}

	return user, nil
}

// GetAllUsers mengambil semua user dengan role "kasir" saja (untuk admin)
func (s *authService) GetAllUsers() ([]user_model.User, error) {
	var users []user_model.User
	if err := config.DB.Where("role = ?", "kasir").Find(&users).Error; err != nil {
		return nil, ErrGetUsersFailed
	}
	return users, nil
}

// DeleteUser menghapus user berdasarkan ID
func (s *authService) DeleteUser(userID string) error {
	var user user_model.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return ErrDeleteUserFailed
	}

	// Jangan izinkan hapus admin
	if user.Role == "admin" {
		return errors.New("Tidak dapat menghapus user admin")
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		return ErrDeleteUserFailed
	}

	return nil
}


