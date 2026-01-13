package services

import (
	"errors"
	"pos-go/config"
	"pos-go/dto"
	category_model "pos-go/models/category_model"

	"gorm.io/gorm"
)

// Sentinel errors
var (
	ErrCategoryNameExists  = errors.New("Nama category sudah digunakan")
	ErrCreateCategoryFailed = errors.New("Gagal membuat category")
)

type CategoryService interface {
	CreateCategory(input dto.CreateCategoryDTO) (category_model.Category, error)
	GetAllCategories() ([]category_model.Category, error)
}

type categoryService struct{}

func NewCategoryService() CategoryService {
	return &categoryService{}
}

func (s *categoryService) CreateCategory(input dto.CreateCategoryDTO) (category_model.Category, error) {
	// Validasi: cek apakah nama category sudah ada
	var existingCategory category_model.Category
	if err := config.DB.Where("name = ?", input.Name).First(&existingCategory).Error; err == nil {
		// Category dengan nama ini sudah ada
		return category_model.Category{}, ErrCategoryNameExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Error lain saat query
		return category_model.Category{}, ErrCreateCategoryFailed
	}

	// Buat category baru
	category := category_model.Category{
		Name: input.Name,
	}

	// Simpan ke database
	if err := config.DB.Create(&category).Error; err != nil {
		return category_model.Category{}, ErrCreateCategoryFailed
	}

	return category, nil
}

func (s *categoryService) GetAllCategories() ([]category_model.Category, error) {
	var categories []category_model.Category
	
	if err := config.DB.Find(&categories).Error; err != nil {
		return nil, ErrCreateCategoryFailed 
	}
	
	return categories, nil
}
