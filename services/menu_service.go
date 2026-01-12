package services

import (
	"errors"
	"pos-go/config"
	"pos-go/dto"
	category_model "pos-go/models/category_model"
	menu_model "pos-go/models/menu_model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Sentinel errors
var (
	ErrCategoryNotFound = errors.New("Category tidak ditemukan")
	ErrCreateMenuFailed = errors.New("Gagal membuat menu")
)

type MenuService interface {
	CreateMenu(input dto.CreateMenuDTO) (menu_model.Menu, error)
}

type menuService struct{}

func NewMenuService() MenuService {
	return &menuService{}
}

func (s *menuService) CreateMenu(input dto.CreateMenuDTO) (menu_model.Menu, error) {
	// Validasi: cek apakah category exists
	categoryID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		return menu_model.Menu{}, ErrCategoryNotFound
	}

	var category category_model.Category
	if err := config.DB.Where("id = ?", categoryID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return menu_model.Menu{}, ErrCategoryNotFound
		}
		return menu_model.Menu{}, ErrCreateMenuFailed
	}

	// Buat menu
	menu := menu_model.Menu{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Image:       input.Image,
		IsAvailable: input.IsAvailable,
		CategoryID:  categoryID,
	}

	// Simpan ke database
	if err := config.DB.Create(&menu).Error; err != nil {
		return menu_model.Menu{}, ErrCreateMenuFailed
	}

	// Preload category untuk response
	if err := config.DB.Preload("Category").First(&menu, menu.ID).Error; err != nil {
		return menu_model.Menu{}, ErrCreateMenuFailed
	}

	return menu, nil
}
