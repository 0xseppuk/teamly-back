package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `gorm:"primaryKey" json:"id"`
	Email         string    `gorm:"not null;uniqueIndex" json:"email"`
	PasswordHash  string    `gorm:"not null" json:"-"` // "-" скрывает из JSON
	Nickname      string    `gorm:"not null;uniqueIndex" json:"nickname"`
	AvatarURL     *string   `gorm:"" json:"avatar_url,omitempty"`
	Discord       *string   `gorm:"" json:"discord,omitempty"`
	Telegram      *string   `gorm:"" json:"telegram,omitempty"`
	CountryCode   *string   `gorm:"size:2" json:"country_code,omitempty"` // ISO код страны (опционально)
	Country       *Country  `gorm:"foreignKey:CountryCode;references:Code" json:"country,omitempty"`
	Description   *string   `gorm:"type:text" json:"description,omitempty"`                // Описание профиля
	BirthDate     *Date     `gorm:"type:date" json:"birth_date,omitempty"`                 // Дата рождения
	Gender        *string   `gorm:"size:20" json:"gender,omitempty"`                       // Пол (male, female, other)
	Languages     []string  `gorm:"type:jsonb;serializer:json" json:"languages,omitempty"` // Языки, которыми владеет пользователь
	LikesCount    int       `gorm:"default:0" json:"likes_count"`
	DislikesCount int       `gorm:"default:0" json:"dislikes_count"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type AuthRequest struct {
	Email          string `json:"email"`
	Nickname       string `json:"nickname"`
	Password       string `json:"password"`
	RecaptchaToken string `json:"recaptchaToken"`
}

type LoginRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	RecaptchaToken string `json:"recaptchaToken"`
}
