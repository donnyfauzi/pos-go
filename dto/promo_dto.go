package dto

import (
	"time"
)

// CreatePromoDTO untuk create promo baru
type CreatePromoDTO struct {
	Code        string    `json:"code" binding:"required,min=3,max=50"`
	Name        string    `json:"name" binding:"required,max=255"`
	Description string    `json:"description"`
	Type        string    `json:"type" binding:"required,oneof=percentage fixed"`
	Value       float64   `json:"value" binding:"required,min=0"`
	MinPurchase float64   `json:"min_purchase"`
	MaxDiscount float64   `json:"max_discount"`
	UsageLimit  int       `json:"usage_limit"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	IsActive    bool      `json:"is_active"`
}

// UpdatePromoDTO untuk update promo
type UpdatePromoDTO struct {
	Code        string    `json:"code" binding:"omitempty,min=3,max=50"`
	Name        string    `json:"name" binding:"omitempty,max=255"`
	Description string    `json:"description"`
	Type        string    `json:"type" binding:"omitempty,oneof=percentage fixed"`
	Value       float64   `json:"value" binding:"omitempty,min=0"`
	MinPurchase float64   `json:"min_purchase"`
	MaxDiscount float64   `json:"max_discount"`
	UsageLimit  int       `json:"usage_limit"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsActive    bool      `json:"is_active"`
}

// ValidatePromoDTO untuk validate promo code saat checkout
type ValidatePromoDTO struct {
	Code     string  `json:"code" binding:"required"`
	Subtotal float64 `json:"subtotal" binding:"required,min=0"`
}
