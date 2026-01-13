package dto

type CreateMenuDTO struct {
	Name        string  `form:"name" binding:"required"`
	Description string  `form:"description"`
	Price       float64 `form:"price" binding:"required,min=0"`
	Image       string  // Tidak di-bind, di-set manual
	IsAvailable bool    // Tidak di-bind, di-set manual
	CategoryID  string  `form:"category_id" binding:"required,uuid"`
}
