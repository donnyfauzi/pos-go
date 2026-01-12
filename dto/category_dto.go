package dto

type CreateCategoryDTO struct {
	Name string `json:"name" binding:"required"`
}
