package database

import (
	"log"
	"pos-go/config"
	user_model "pos-go/models/user_model"

	"golang.org/x/crypto/bcrypt"
)

func SeedAdmin() {
	// 1. CEK apakah admin sudah ada
	var admin user_model.User
	if err := config.DB.Where("email = ?", "adminresto@gmail.com").First(&admin).Error; err == nil {
		// Admin sudah ada → SKIP (tidak buat lagi)
		log.Println("Admin sudah ada, skip seeding")
		return
	}

	// 2. Admin belum ada → BUAT admin baru
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Gagal hash password admin:", err)
	}

	// Buat admin
	admin = user_model.User{
		Name:     "Admin Utama",
		Email:    "adminresto@gmail.com",
		Password: string(hashedPassword),
		Role:     "admin",
	}

	if err := config.DB.Create(&admin).Error; err != nil {
		log.Fatal("Gagal create admin:", err)
	}

	log.Println("Admin berhasil dibuat: adminresto@gmail.com / admin123")
}
