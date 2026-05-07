package main

import (
	"log"

	"absensi-app/internal/config"
	"absensi-app/internal/database"
	"absensi-app/internal/model"
	"absensi-app/internal/repository"
	"absensi-app/internal/service"
)

func main() {
	log.Println("Starting database seeding...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.InitDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations first
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Create seed users
	users := []struct {
		username string
		password string
		fullName string
		role     string
	}{
		{"admin", "admin123", "Administrator", "admin"},
		{"user1", "password123", "Budi Santoso", "employee"},
		{"user2", "password123", "Siti Nurhaliza", "employee"},
		{"user3", "password123", "Ahmad Wijaya", "employee"},
	}

	for _, u := range users {
		// Check if user already exists
		existing, _ := userRepo.FindByUsername(u.username)
		if existing != nil {
			log.Printf("User %s already exists, skipping...", u.username)
			continue
		}

		// Hash password
		passwordHash, err := service.HashPassword(u.password)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", u.username, err)
			continue
		}

		// Create user
		user := &model.User{
			Username:     u.username,
			PasswordHash: passwordHash,
			FullName:     u.fullName,
			Role:         u.role,
			IsActive:     true,
		}

		if err := userRepo.Create(user); err != nil {
			log.Printf("Failed to create user %s: %v", u.username, err)
			continue
		}

		log.Printf("✓ Created user: %s (password: %s)", u.username, u.password)
	}

	log.Println("\nSeeding completed!")
	log.Println("\nDefault credentials:")
	log.Println("  Admin - username: admin, password: admin123")
	log.Println("  User  - username: user1, password: password123")
}
