package services

import (
	"errors"
	"fmt"
	"pos-go/config"
	"pos-go/dto"
	promo_model "pos-go/models/promo_model"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Sentinel errors
var (
	ErrPromoNotFound     = errors.New("Promo tidak ditemukan")
	ErrPromoCodeExists   = errors.New("Kode promo sudah digunakan")
	ErrPromoInactive     = errors.New("Promo tidak aktif")
	ErrPromoExpired      = errors.New("Promo sudah kadaluarsa")
	ErrPromoUsageLimitReached = errors.New("Promo sudah habis")
	ErrMinPurchaseNotMet = errors.New("Minimum pembelian tidak terpenuhi")
	ErrCreatePromoFailed = errors.New("Gagal membuat promo")
	ErrUpdatePromoFailed = errors.New("Gagal mengupdate promo")
	ErrDeletePromoFailed = errors.New("Gagal menghapus promo")
	ErrGetPromosFailed   = errors.New("Gagal mengambil daftar promo")
)

type PromoService interface {
	CreatePromo(input dto.CreatePromoDTO) (promo_model.Promo, error)
	UpdatePromo(promoID string, input dto.UpdatePromoDTO) (promo_model.Promo, error)
	DeletePromo(promoID string) error
	GetAllPromos() ([]promo_model.Promo, error)
	GetPromoByID(promoID string) (promo_model.Promo, error)
	ValidatePromo(code string, subtotal float64) (promo_model.Promo, float64, error)
}

type promoService struct{}

func NewPromoService() PromoService {
	return &promoService{}
}


// CreatePromo membuat promo baru
func (s *promoService) CreatePromo(input dto.CreatePromoDTO) (promo_model.Promo, error) {
	// Validasi: cek apakah kode promo sudah ada (case insensitive)
	var existingPromo promo_model.Promo
	if err := config.DB.Where("LOWER(code) = LOWER(?)", input.Code).First(&existingPromo).Error; err == nil {
		return promo_model.Promo{}, ErrPromoCodeExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return promo_model.Promo{}, ErrCreatePromoFailed
	}

	// Convert code ke uppercase untuk konsistensi
	input.Code = strings.ToUpper(input.Code)

	// Buat promo
	promo := promo_model.Promo{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
		Type:        input.Type,
		Value:       input.Value,
		MinPurchase: input.MinPurchase,
		MaxDiscount: input.MaxDiscount,
		UsageLimit:  input.UsageLimit,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		IsActive:    input.IsActive,
	}

	// Simpan ke database
	if err := config.DB.Create(&promo).Error; err != nil {
		return promo_model.Promo{}, ErrCreatePromoFailed
	}

	return promo, nil
}

// GetAllPromos mengambil semua promo
func (s *promoService) GetAllPromos() ([]promo_model.Promo, error) {
	var promos []promo_model.Promo

	if err := config.DB.Order("created_at DESC").Find(&promos).Error; err != nil {
		return nil, ErrGetPromosFailed
	}

	return promos, nil
}

// GetPromoByID mengambil promo berdasarkan ID
func (s *promoService) GetPromoByID(promoID string) (promo_model.Promo, error) {
	id, err := uuid.Parse(promoID)
	if err != nil {
		return promo_model.Promo{}, ErrPromoNotFound
	}

	var promo promo_model.Promo
	if err := config.DB.Where("id = ?", id).First(&promo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return promo_model.Promo{}, ErrPromoNotFound
		}
		return promo_model.Promo{}, ErrGetPromosFailed
	}

	return promo, nil
}

// UpdatePromo mengupdate promo (partial update)
func (s *promoService) UpdatePromo(promoID string, input dto.UpdatePromoDTO) (promo_model.Promo, error) {
	id, err := uuid.Parse(promoID)
	if err != nil {
		return promo_model.Promo{}, ErrPromoNotFound
	}

	// Cek apakah promo exists
	var promo promo_model.Promo
	if err := config.DB.Where("id = ?", id).First(&promo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return promo_model.Promo{}, ErrPromoNotFound
		}
		return promo_model.Promo{}, ErrUpdatePromoFailed
	}

	// Partial update: hanya update field yang dikirim
	if input.Code != "" {
		// Validasi: cek apakah kode sudah digunakan oleh promo lain
		var existingPromo promo_model.Promo
		if err := config.DB.Where("LOWER(code) = LOWER(?) AND id != ?", input.Code, id).First(&existingPromo).Error; err == nil {
			return promo_model.Promo{}, ErrPromoCodeExists
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return promo_model.Promo{}, ErrUpdatePromoFailed
		}
		promo.Code = strings.ToUpper(input.Code)
	}

	if input.Name != "" {
		promo.Name = input.Name
	}

	promo.Description = input.Description // Boleh kosong

	if input.Type != "" {
		promo.Type = input.Type
	}
	if input.Value > 0 {
		promo.Value = input.Value
	}

	// Update numeric fields (0 = unlimited/disabled)
	promo.MinPurchase = input.MinPurchase
	promo.MaxDiscount = input.MaxDiscount
	promo.UsageLimit = input.UsageLimit

	if !input.StartDate.IsZero() {
		promo.StartDate = input.StartDate
	}

	if !input.EndDate.IsZero() {
		promo.EndDate = input.EndDate
	}

	promo.IsActive = input.IsActive

	// Simpan perubahan
	if err := config.DB.Save(&promo).Error; err != nil {
		return promo_model.Promo{}, ErrUpdatePromoFailed
	}

	return promo, nil
}

// DeletePromo menghapus promo (soft delete)
func (s *promoService) DeletePromo(promoID string) error {
	id, err := uuid.Parse(promoID)
	if err != nil {
		return ErrPromoNotFound
	}

	// Cek apakah promo exists
	var promo promo_model.Promo
	if err := config.DB.Where("id = ?", id).First(&promo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPromoNotFound
		}
		return ErrDeletePromoFailed
	}

	// Soft delete
	if err := config.DB.Delete(&promo).Error; err != nil {
		return ErrDeletePromoFailed
	}

	return nil
}

// ValidatePromo memvalidasi kode promo dan menghitung discount
func (s *promoService) ValidatePromo(code string, subtotal float64) (promo_model.Promo, float64, error) {
	var promo promo_model.Promo

	// Find promo by code (case insensitive)
	if err := config.DB.Where("LOWER(code) = LOWER(?)", code).First(&promo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return promo_model.Promo{}, 0, ErrPromoNotFound
		}
		return promo_model.Promo{}, 0, errors.New("Gagal memvalidasi promo")
	}

	// Check: is_active
	if !promo.IsActive {
		return promo_model.Promo{}, 0, ErrPromoInactive
	}

	// Check: date range
	now := time.Now()
	if now.Before(promo.StartDate) || now.After(promo.EndDate) {
		return promo_model.Promo{}, 0, ErrPromoExpired
	}

	// Check: usage limit (0 = unlimited)
	if promo.UsageLimit > 0 && promo.UsageCount >= promo.UsageLimit {
		return promo_model.Promo{}, 0, ErrPromoUsageLimitReached
	}

	// Check: min purchase
	if subtotal < promo.MinPurchase {
		return promo_model.Promo{}, 0, fmt.Errorf("Minimum pembelian Rp %.0f", promo.MinPurchase)
	}

	// Calculate discount
	var discount float64
	if promo.Type == "percentage" {
		discount = subtotal * (promo.Value / 100)
	} else { // fixed
		discount = promo.Value
	}

	// Apply max discount (0 = unlimited)
	if promo.MaxDiscount > 0 && discount > promo.MaxDiscount {
		discount = promo.MaxDiscount
	}

	// Ensure discount tidak lebih dari subtotal
	if discount > subtotal {
		discount = subtotal
	}

	return promo, discount, nil
}