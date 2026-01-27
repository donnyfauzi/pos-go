package dto

type CreateMenuDTO struct {
	Name        string  `form:"name" binding:"required"`
	Description string  `form:"description"`
	Price       float64 `form:"price" binding:"required,min=0"`
	Image       string  // Tidak di-bind, di-set manual
	IsAvailable bool    // Tidak di-bind, di-set manual
	CategoryID  string  `form:"category_id" binding:"required,uuid"`
}

type UpdateMenuDTO struct {
	Name        string  `form:"name" binding:"omitempty"`
	Description string  `form:"description"`
	Price       float64 `form:"price" binding:"omitempty,min=0"`
	Image       string  // Tidak di-bind, di-set manual (opsional saat update)
	IsAvailable bool    // Tidak di-bind, di-set manual
	CategoryID  string  `form:"category_id" binding:"omitempty,uuid"`
}
