package controllers

import (
	"errors"
	"pos-go/dto"
	"pos-go/services"
	"pos-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var authService = services.NewAuthService()

func Register(c *gin.Context) {
	var input dto.RegisterDTO

	// Validasi request body (HTTP level)
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsMap := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()

				switch fe.Tag() {
				case "required":
					errorsMap[field] = "Field wajib diisi"
				case "email":
					errorsMap[field] = "Format email tidak valid"
				case "min":
					errorsMap[field] = field + " minimal harus " + fe.Param() + " karakter"
				case "oneof":
					errorsMap[field] = field + " harus salah satu dari: " + fe.Param()
				default:
					errorsMap[field] = "Field tidak valid"
				}
			}

			utils.ErrorResponseBadRequest(c, "Validasi gagal", errorsMap)
			return
		}

		// Kalau bukan error dari validator (JSON tidak valid, dll)
		utils.ErrorResponseBadRequest(c, "Format request tidak valid", nil)
		return
	}

	// Jalankan business logic via service
	user, err := authService.Register(input)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			utils.ErrorResponseConflict(c, "Email sudah terdaftar, silahkan gunakan email lain")
			return
		}

		if errors.Is(err, services.ErrHashPasswordFailed) {
			utils.ErrorResponseBadRequest(c, "Gagal memproses password", nil)
			return
		}

		if errors.Is(err, services.ErrCreateUserFailed) {
			utils.ErrorResponseInternal(c, "Gagal menyimpan user")
			return
		}

		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	// Sukses register
	utils.SuccessResponseCreated(c, "User berhasil dibuat", user)
}

func Login(c *gin.Context) {
	var input dto.LoginDTO

	// Validasi request body (HTTP level)
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsMap := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()

				switch fe.Tag() {
				case "required":
					errorsMap[field] = "Field wajib diisi"
				case "email":
					errorsMap[field] = "Format email tidak valid"
				default:
					errorsMap[field] = "Field tidak valid"
				}
			}

			utils.ErrorResponseBadRequest(c, "Validasi gagal", errorsMap)
			return
		}

		// Kalau bukan error dari validator (JSON tidak valid, dll)
		utils.ErrorResponseBadRequest(c, "Format request tidak valid", nil)
		return
	}

	// Jalankan business logic via service
	user, token, err := authService.Login(input)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.ErrorResponseUnauthorized(c, "Email atau password salah, silahkan coba lagi")
			return
		}

		if errors.Is(err, services.ErrCreateUserFailed) {
			utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
			return
		}

		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	// Set token di HTTP-only cookie (Hybrid approach)
	c.SetCookie(
		"token", // name
		token,   // value
		3600,    // max age (1 jam dalam detik)
		"/",     // path
		"",      // domain (kosong = current domain)
		false,   // secure (set true jika pakai HTTPS)
		true,    // httpOnly (true = tidak bisa diakses JavaScript)
	)

	// Sukses login - return user (token sudah di cookie)
	utils.SuccessResponseOK(c, "Login berhasil", gin.H{
		"user": user,
	})
}

func ChangePassword(c *gin.Context) {
	var input dto.ChangePasswordDTO

	// Validasi request body (HTTP level)
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsMap := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()

				switch fe.Tag() {
				case "required":
					errorsMap[field] = "Field wajib diisi"
				case "min":
					errorsMap[field] = field + " minimal harus " + fe.Param() + " karakter"
				default:
					errorsMap[field] = "Field tidak valid"
				}
			}

			utils.ErrorResponseBadRequest(c, "Validasi gagal", errorsMap)
			return
		}

		utils.ErrorResponseBadRequest(c, "Format request tidak valid", nil)
		return
	}

	// Ambil user_id dari context (dari AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponseUnauthorized(c, "User ID tidak ditemukan")
		return
	}

	// Jalankan business logic via service
	err := authService.ChangePassword(userID.(string), input)
	if err != nil {
		if errors.Is(err, services.ErrInvalidOldPassword) {
			utils.ErrorResponseUnauthorized(c, "Password lama salah")
			return
		}

		if errors.Is(err, services.ErrHashPasswordFailed) {
			utils.ErrorResponseInternal(c, "Gagal memproses password baru")
			return
		}

		if errors.Is(err, services.ErrCreateUserFailed) {
			utils.ErrorResponseInternal(c, "Gagal mengubah password")
			return
		}

		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	// Sukses change password
	utils.SuccessResponseOK(c, "Password berhasil diubah", nil)
}

func GetCurrentUser(c *gin.Context) {
	// Ambil user_id dari context (dari AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponseUnauthorized(c, "User ID tidak ditemukan")
		return
	}

	// Get user dari database
	user, err := authService.GetCurrentUser(userID.(string))
	if err != nil {
		utils.ErrorResponseUnauthorized(c, "User tidak ditemukan")
		return
	}

	// Return user
	utils.SuccessResponseOK(c, "User berhasil ditemukan", user)
}

func GetAllUsers(c *gin.Context) {
	users, err := authService.GetAllUsers()
	if err != nil {
		if errors.Is(err, services.ErrGetUsersFailed) {
			utils.ErrorResponseInternal(c, "Gagal mengambil daftar user")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	utils.SuccessResponseOK(c, "Berhasil mengambil daftar user", users)
}

func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	err := authService.DeleteUser(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			utils.ErrorResponseNotFound(c, "User tidak ditemukan")
			return
		}
		if err.Error() == "Tidak dapat menghapus user admin" {
			utils.ErrorResponseBadRequest(c, err.Error(), nil)
			return
		}
		if errors.Is(err, services.ErrDeleteUserFailed) {
			utils.ErrorResponseInternal(c, "Gagal menghapus user")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	utils.SuccessResponseOK(c, "User berhasil dihapus", nil)
}
