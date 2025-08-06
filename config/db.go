package config

import (
	"fmt"
	"log"
	"os"

	user_model "pos-go/models/user_model"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

	// koneksi .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Gagal memuat file .env")
	}

	// Ambil variabel dari .env
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Format DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)

	// koneksi ke database
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	log.Println("Berhasil terhubung ke database PostgreSQL")

	// AutoMigrate tabel user
	err = database.AutoMigrate(&user_model.User{})
	if err != nil {
		log.Fatal("Migrasi gagal:", err)
	}

	log.Println("Migrasi tabel berhasil")

	DB = database
}
