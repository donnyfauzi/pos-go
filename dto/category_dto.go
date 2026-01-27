package dto

type CreateCategoryDTO struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCategoryDTO struct {
	Name string `json:"name" binding:"omitempty"`
}
