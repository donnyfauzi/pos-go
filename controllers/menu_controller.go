package controllers

import (
	"pos-go/dto"
	"pos-go/services"
	"pos-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var menuService services.MenuService = services.NewMenuService()

func CreateMenu(c *gin.Context) {
	// Bind form data (multipart/form-data)
	var input dto.CreateMenuDTO
	
	// Bind form fields
	if err := c.ShouldBind(&input); err != nil {
		errorsMap := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()

				switch fe.Tag() {
				case "required":
					errorsMap[field] = "Field wajib diisi"
				case "min":
					errorsMap[field] = field + " minimal harus " + fe.Param()
				case "uuid":
					errorsMap[field] = "Format " + field + " tidak valid"
				default:
					errorsMap[field] = "Field tidak valid"
				}
			}

			utils.ErrorResponseBadRequest(c, "Validasi gagal", errorsMap)
			return
		}

		// Kalau bukan error dari validator (format data tidak valid, dll)
		utils.ErrorResponseBadRequest(c, "Format data tidak valid", nil)
		return
	}

	// Validasi image wajib
	_, err := c.FormFile("image")
	if err != nil {
		utils.ErrorResponseBadRequest(c, "Image wajib diisi", nil)
		return
	}

	// Upload file
	imagePath, err := utils.SaveUploadedFile(c, "image")
	if err != nil {
		utils.ErrorResponseBadRequest(c, err.Error(), nil)
		return
	}
	input.Image = imagePath

	// Create menu via service
	menu, err := menuService.CreateMenu(input)
	if err != nil {
		// Handle sentinel errors
		if err == services.ErrCategoryNotFound {
			utils.ErrorResponseBadRequest(c, "Category tidak ditemukan", nil)
			return
		}
		if err == services.ErrCreateMenuFailed {
			utils.ErrorResponseInternal(c, "Gagal membuat menu")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	utils.SuccessResponseCreated(c, "Menu berhasil dibuat", menu)
}
