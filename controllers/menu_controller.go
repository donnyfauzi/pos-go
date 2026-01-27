package controllers

import (
	"errors"
	"pos-go/dto"
	"pos-go/services"
	"pos-go/utils"
	"strconv"

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

	// Parse is_available manually untuk memastikan boolean di-parse dengan benar
	// Default: false (jika tidak ada atau kosong)
	input.IsAvailable = false
	isAvailableStr := c.PostForm("is_available")
	if isAvailableStr != "" {
		if isAvailable, err := strconv.ParseBool(isAvailableStr); err == nil {
			input.IsAvailable = isAvailable
		}
	}

	// Create menu via service
	menu, err := menuService.CreateMenu(input)
	if err != nil {
		// Handle sentinel errors
		if errors.Is(err, services.ErrCategoryNotFound) {
			utils.ErrorResponseBadRequest(c, "Category tidak ditemukan", nil)
			return
		}
		if errors.Is(err, services.ErrMenuNameExists) {
			utils.ErrorResponseBadRequest(c, "Nama menu sudah digunakan", nil)
			return
		}
		if errors.Is(err, services.ErrCreateMenuFailed) {
			utils.ErrorResponseInternal(c, "Gagal membuat menu")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	utils.SuccessResponseCreated(c, "Menu berhasil dibuat", menu)
}

func GetAllMenus(c *gin.Context) {
	menus, err := menuService.GetAllMenus()
	if err != nil {
		if errors.Is(err, services.ErrGetMenusFailed) {
			utils.ErrorResponseInternal(c, "Gagal mengambil daftar menu")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	utils.SuccessResponseOK(c, "Berhasil mengambil daftar menu", menus)
}

func GetPublicMenus(c *gin.Context) {
	menus, err := menuService.GetPublicMenus()
	if err != nil {
		if errors.Is(err, services.ErrGetMenusFailed) {
			utils.ErrorResponseInternal(c, "Gagal mengambil daftar menu")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	utils.SuccessResponseOK(c, "Berhasil mengambil daftar menu", menus)
}

func UpdateMenu(c *gin.Context) {
	menuID := c.Param("id")

	// Bind form data (multipart/form-data)
	var input dto.UpdateMenuDTO

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

	// Image optional (jika tidak ada, gunakan image yang lama)
	_, err := c.FormFile("image")
	if err == nil {
		// Ada image baru, upload
		imagePath, err := utils.SaveUploadedFile(c, "image")
		if err != nil {
			utils.ErrorResponseBadRequest(c, err.Error(), nil)
			return
		}
		input.Image = imagePath
	} else {
		// Tidak ada image baru, set kosong (service akan menggunakan image lama)
		input.Image = ""
	}

	// Parse is_available manually untuk memastikan boolean di-parse dengan benar
	input.IsAvailable = false
	isAvailableStr := c.PostForm("is_available")
	if isAvailableStr != "" {
		if isAvailable, err := strconv.ParseBool(isAvailableStr); err == nil {
			input.IsAvailable = isAvailable
		}
	}

	// Update menu via service
	menu, err := menuService.UpdateMenu(menuID, input)
	if err != nil {
		// Handle sentinel errors
		if errors.Is(err, services.ErrMenuNotFound) {
			utils.ErrorResponseBadRequest(c, "Menu tidak ditemukan", nil)
			return
		}
		if errors.Is(err, services.ErrCategoryNotFound) {
			utils.ErrorResponseBadRequest(c, "Category tidak ditemukan", nil)
			return
		}
		if errors.Is(err, services.ErrMenuNameExists) {
			utils.ErrorResponseBadRequest(c, "Nama menu sudah digunakan", nil)
			return
		}
		if errors.Is(err, services.ErrUpdateMenuFailed) {
			utils.ErrorResponseInternal(c, "Gagal mengupdate menu")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	utils.SuccessResponseOK(c, "Menu berhasil diperbarui", menu)
}

func DeleteMenu(c *gin.Context) {
	menuID := c.Param("id")

	// Delete menu via service
	err := menuService.DeleteMenu(menuID)
	if err != nil {
		// Handle sentinel errors
		if errors.Is(err, services.ErrMenuNotFound) {
			utils.ErrorResponseBadRequest(c, "Menu tidak ditemukan", nil)
			return
		}
		if errors.Is(err, services.ErrDeleteMenuFailed) {
			utils.ErrorResponseInternal(c, "Gagal menghapus menu")
			return
		}
		utils.ErrorResponseInternal(c, "Terjadi kesalahan pada server")
		return
	}

	utils.SuccessResponseOK(c, "Menu berhasil dihapus", nil)
}
