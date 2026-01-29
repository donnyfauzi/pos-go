package controllers

import (
	"errors"
	"pos-go/dto"
	"pos-go/services"
	"pos-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var promoService services.PromoService = services.NewPromoService()

func CreatePromo(c *gin.Context) {
	var input dto.CreatePromoDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		errorsMap := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()
				switch fe.Tag() {
				case "required":
					errorsMap[field] = "Field wajib diisi"
				case "min":
					errorsMap[field] = "Nilai minimal " + fe.Param()
				case "max":
					errorsMap[field] = "Nilai maksimal " + fe.Param()
				case "oneof":
					errorsMap[field] = "Pilihan tidak valid"
				default:
					errorsMap[field] = "Field tidak valid"
				}
			}
			utils.ErrorResponseBadRequest(c, "Validasi gagal", errorsMap)
			return
		}

		utils.ErrorResponseBadRequest(c, "Format data tidak valid", nil)
		return
	}

	promo, err := promoService.CreatePromo(input)
	if err != nil {
		if errors.Is(err, services.ErrPromoCodeExists) {
			utils.ErrorResponseBadRequest(c, "Kode promo sudah digunakan", nil)
			return
		}
		utils.ErrorResponseInternal(c, "Gagal membuat promo")
		return
	}

	utils.SuccessResponseCreated(c, "Promo berhasil dibuat", promo)
}

func GetAllPromos(c *gin.Context) {
	promos, err := promoService.GetAllPromos()
	if err != nil {
		utils.ErrorResponseInternal(c, "Gagal mengambil daftar promo")
		return
	}

	utils.SuccessResponseOK(c, "Berhasil mengambil daftar promo", promos)
}

func GetPromoByID(c *gin.Context) {
	promoID := c.Param("id")

	promo, err := promoService.GetPromoByID(promoID)
	if err != nil {
		if errors.Is(err, services.ErrPromoNotFound) {
			utils.ErrorResponseNotFound(c, "Promo tidak ditemukan")
			return
		}
		utils.ErrorResponseInternal(c, "Gagal mengambil data promo")
		return
	}

	utils.SuccessResponseOK(c, "Berhasil mengambil data promo", promo)
}

func UpdatePromo(c *gin.Context) {
	promoID := c.Param("id")
	var input dto.UpdatePromoDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		errorsMap := make(map[string]string)

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.Field()
				switch fe.Tag() {
				case "min":
					errorsMap[field] = "Nilai minimal " + fe.Param()
				case "max":
					errorsMap[field] = "Nilai maksimal " + fe.Param()
				case "oneof":
					errorsMap[field] = "Pilihan tidak valid"
				default:
					errorsMap[field] = "Field tidak valid"
				}
			}
			utils.ErrorResponseBadRequest(c, "Validasi gagal", errorsMap)
			return
		}

		utils.ErrorResponseBadRequest(c, "Format data tidak valid", nil)
		return
	}

	promo, err := promoService.UpdatePromo(promoID, input)
	if err != nil {
		if errors.Is(err, services.ErrPromoNotFound) {
			utils.ErrorResponseNotFound(c, "Promo tidak ditemukan")
			return
		}
		if errors.Is(err, services.ErrPromoCodeExists) {
			utils.ErrorResponseBadRequest(c, "Kode promo sudah digunakan", nil)
			return
		}
		utils.ErrorResponseInternal(c, "Gagal mengupdate promo")
		return
	}

	utils.SuccessResponseOK(c, "Promo berhasil diupdate", promo)
}

func DeletePromo(c *gin.Context) {
	promoID := c.Param("id")

	err := promoService.DeletePromo(promoID)
	if err != nil {
		if errors.Is(err, services.ErrPromoNotFound) {
			utils.ErrorResponseNotFound(c, "Promo tidak ditemukan")
			return
		}
		utils.ErrorResponseInternal(c, "Gagal menghapus promo")
		return
	}

	utils.SuccessResponseOK(c, "Promo berhasil dihapus", nil)
}

// GetActivePromos - Public endpoint untuk customer melihat promo aktif
func GetActivePromos(c *gin.Context) {
	promos, err := promoService.GetActivePromos()
	if err != nil {
		utils.ErrorResponseInternal(c, "Gagal mengambil daftar promo aktif")
		return
	}

	utils.SuccessResponseOK(c, "Berhasil mengambil daftar promo aktif", promos)
}

// ValidatePromo - Public endpoint untuk customer validate promo saat checkout
func ValidatePromo(c *gin.Context) {
	var input dto.ValidatePromoDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponseBadRequest(c, "Format data tidak valid", nil)
		return
	}

	promo, discount, err := promoService.ValidatePromo(input.Code, input.Subtotal)
	if err != nil {
		utils.ErrorResponseBadRequest(c, err.Error(), nil)
		return
	}

	response := gin.H{
		"promo":        promo,
		"discount":     discount,
		"final_amount": input.Subtotal - discount,
	}

	utils.SuccessResponseOK(c, "Promo valid", response)
}