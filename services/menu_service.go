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
	ErrMenuNotFound     = errors.New("Menu tidak ditemukan")
	ErrMenuNameExists   = errors.New("Nama menu sudah digunakan")
	ErrCreateMenuFailed = errors.New("Gagal membuat menu")
	ErrUpdateMenuFailed = errors.New("Gagal mengupdate menu")
	ErrDeleteMenuFailed = errors.New("Gagal menghapus menu")
	ErrGetMenusFailed   = errors.New("Gagal mengambil daftar menu")
)

type MenuService interface {
	CreateMenu(input dto.CreateMenuDTO) (menu_model.Menu, error)
	UpdateMenu(menuID string, input dto.CreateMenuDTO) (menu_model.Menu, error)
	DeleteMenu(menuID string) error
	GetAllMenus() ([]menu_model.Menu, error) //admin
	GetPublicMenus() ([]menu_model.Menu, error)
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

	// Validasi: cek apakah nama menu sudah ada
	var existingMenu menu_model.Menu
	if err := config.DB.Where("name = ?", input.Name).First(&existingMenu).Error; err == nil {
		// Menu dengan nama ini sudah ada
		return menu_model.Menu{}, ErrMenuNameExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Error lain saat query
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

// Get semua menu untuk admin dashboard
func (s *menuService) GetAllMenus() ([]menu_model.Menu, error) {
	var menus []menu_model.Menu

	// Preload Category untuk mendapatkan informasi kategori
	if err := config.DB.Preload("Category").Find(&menus).Error; err != nil {
		return nil, ErrGetMenusFailed
	}

	return menus, nil
}

// Get menu yang tersedia untuk customer (is_available = true)
func (s *menuService) GetPublicMenus() ([]menu_model.Menu, error) {
	var menus []menu_model.Menu

	// Filter hanya menu yang is_available = true
	// Preload Category untuk mendapatkan informasi kategori
	if err := config.DB.Preload("Category").Where("is_available = ?", true).Find(&menus).Error; err != nil {
		return nil, ErrGetMenusFailed
	}

	return menus, nil
}

// UpdateMenu mengupdate menu berdasarkan ID
func (s *menuService) UpdateMenu(menuID string, input dto.CreateMenuDTO) (menu_model.Menu, error) {
	// Parse menu ID
	id, err := uuid.Parse(menuID)
	if err != nil {
		return menu_model.Menu{}, ErrMenuNotFound
	}

	// Cek apakah menu exists
	var menu menu_model.Menu
	if err := config.DB.Where("id = ?", id).First(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return menu_model.Menu{}, ErrMenuNotFound
		}
		return menu_model.Menu{}, ErrUpdateMenuFailed
	}

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
		return menu_model.Menu{}, ErrUpdateMenuFailed
	}

	// Validasi: cek apakah nama menu sudah digunakan oleh menu lain (bukan menu yang sedang di-edit)
	var existingMenu menu_model.Menu
	if err := config.DB.Where("name = ? AND id != ?", input.Name, id).First(&existingMenu).Error; err == nil {
		// Menu dengan nama ini sudah ada (bukan menu yang sedang di-edit)
		return menu_model.Menu{}, ErrMenuNameExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Error lain saat query
		return menu_model.Menu{}, ErrUpdateMenuFailed
	}

	// Update fields
	menu.Name = input.Name
	menu.Description = input.Description
	menu.Price = input.Price
	menu.IsAvailable = input.IsAvailable
	menu.CategoryID = categoryID

	// Update image hanya jika ada (input.Image tidak kosong)
	if input.Image != "" {
		menu.Image = input.Image
	}

	// Simpan perubahan ke database
	if err := config.DB.Save(&menu).Error; err != nil {
		return menu_model.Menu{}, ErrUpdateMenuFailed
	}

	// Preload category untuk response
	if err := config.DB.Preload("Category").First(&menu, menu.ID).Error; err != nil {
		return menu_model.Menu{}, ErrUpdateMenuFailed
	}

	return menu, nil
}

// DeleteMenu menghapus menu berdasarkan ID
func (s *menuService) DeleteMenu(menuID string) error {
	// Parse menu ID
	id, err := uuid.Parse(menuID)
	if err != nil {
		return ErrMenuNotFound
	}

	// Cek apakah menu exists
	var menu menu_model.Menu
	if err := config.DB.Where("id = ?", id).First(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return ErrDeleteMenuFailed
	}

	// Hapus menu
	if err := config.DB.Delete(&menu).Error; err != nil {
		return ErrDeleteMenuFailed
	}

	return nil
}
