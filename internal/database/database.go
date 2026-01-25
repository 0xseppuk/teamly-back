package database

import (
	"fmt"
	"log"

	"github.com/duker221/teamly/internal/config"
	"github.com/duker221/teamly/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error
	envError := godotenv.Load()
	if envError != nil {
		log.Fatalf("Failed to load environment variables: %v", envError)
	}

	config.LoadDbConfig()

	log.Printf("DB Config: host=%s port=%s user=%s dbname=%s",
		config.DbConfig.Host, config.DbConfig.Port, config.DbConfig.User, config.DbConfig.DBName)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DbConfig.Host, config.DbConfig.Port, config.DbConfig.User, config.DbConfig.Password, config.DbConfig.DBName)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database with GORM")

	if err := AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}

func AutoMigrate() error {
	log.Println("Running auto migrations...")

	// Сначала мигрируем Country, потому что User зависит от него
	err := DB.AutoMigrate(
		&models.Country{},
		&models.User{},
		&models.Game{},
		&models.GameApplication{},
		&models.ApplicationResponse{},
		&models.Conversation{},
		&models.Message{},
		&models.PasswordResetToken{},
		// &models.Listing{},
		// &models.ListingGame{},
		// &models.Review{},
	)

	if err != nil {
		return fmt.Errorf("auto migration failed: %v", err)
	}

	// Seed countries if table is empty
	if err := SeedCountries(); err != nil {
		return fmt.Errorf("country seeding failed: %v", err)
	}

	// Seed games if table is empty
	if err := SeedGames(); err != nil {
		return fmt.Errorf("game seeding failed: %v", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func SeedCountries() error {
	var count int64
	DB.Model(&models.Country{}).Count(&count)

	// Если страны уже есть, не добавляем заново
	if count > 0 {
		log.Println("Countries already seeded, skipping...")
		return nil
	}

	log.Println("Seeding countries...")
	countries := models.GetCountriesSeed()

	if err := DB.Create(&countries).Error; err != nil {
		return err
	}

	log.Printf("Successfully seeded %d countries", len(countries))
	return nil
}

func SeedGames() error {
	var count int64
	DB.Model(&models.Game{}).Count(&count)

	// Если игры уже есть, не добавляем заново
	if count > 0 {
		log.Println("Games already seeded, skipping...")
		return nil
	}

	log.Println("Seeding games...")
	games := []models.Game{
		{
			Name:     "Counter-Strike 2",
			Icon_url: "https://cdn.cloudflare.steamstatic.com/apps/csgo/images/csgo_react/social/cs2.jpg",
			Slug:     "counter-strike-2",
			IsActive: true,
		},
		{
			Name:     "Dota 2",
			Icon_url: "https://cdn.cloudflare.steamstatic.com/apps/dota2/images/dota_react/global/dota2_logo_symbol.png",
			Slug:     "dota-2",
			IsActive: true,
		},
		{
			Name:     "Valorant",
			Icon_url: "https://images.contentstack.io/v3/assets/bltb6530b271fddd0b1/blt1eb1891a4531c2f9/5eb7cdc0ee88d36e47530fba/V_LOGOMARK_1920x1080_Main.png",
			Slug:     "valorant",
			IsActive: true,
		},
		{
			Name:     "Apex Legends",
			Icon_url: "https://media.contentapi.ea.com/content/dam/apex-legends/common/apex-logo-white.svg",
			Slug:     "apex-legends",
			IsActive: true,
		},
		{
			Name:     "PUBG: BATTLEGROUNDS",
			Icon_url: "https://cdn.cloudflare.steamstatic.com/apps/578080/header.jpg",
			Slug:     "pubg-battlegrounds",
			IsActive: true,
		},
	}

	if err := DB.Create(&games).Error; err != nil {
		return err
	}

	log.Printf("Successfully seeded %d games", len(games))
	return nil
}
