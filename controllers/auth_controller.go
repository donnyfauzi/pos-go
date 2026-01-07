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
			utils.ErrorResponseConflict(c, "Email sudah terdaftar")
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


