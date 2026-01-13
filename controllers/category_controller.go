package controllers

import (
	"pos-go/dto"
	"pos-go/services"
	"pos-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var categoryService services.CategoryService = services.NewCategoryService()

func CreateCategory(c *gin.Context) {
	var input dto.CreateCategoryDTO

	// Bind dan validasi input
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsMap := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()

				switch fe.Tag() {
				case "required":
					errorsMap[field] = "Field wajib diisi"
				default:
					errorsMap[field] = "Field tidak valid"
				}
			}

			utils.ErrorResponseBadRequest(c, "Validasi gagal", errorsMap)
			return
		}

		// Kalau bukan error dari validator (JSON tidak valid, dll)
		utils.ErrorResponseBadRequest(c, "Format data tidak valid", nil)
		return
	}

	// Create category via service
	category, err := categoryService.CreateCategory(input)
	if err != nil {
		// Handle sentinel errors
		if err == services.ErrCategoryNameExists {
			utils.ErrorResponseBadRequest(c, "Nama category sudah digunakan", nil)
			return
		}
		if err == services.ErrCreateCategoryFailed {
			utils.ErrorResponseInternal(c, "Gagal membuat category")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	// Success response
	utils.SuccessResponseCreated(c, "Category berhasil dibuat", category)
}

func GetAllCategories(c *gin.Context) {
	categories, err := categoryService.GetAllCategories()
	if err != nil {
		utils.ErrorResponseInternal(c, "Gagal mengambil daftar category")
		return
	}
	
	utils.SuccessResponseOK(c, "Berhasil mengambil daftar category", categories)
}

