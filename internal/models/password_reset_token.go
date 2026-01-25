package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"not null;index" json:"user_id"`
	Token     string     `gorm:"not null;uniqueIndex;size:64" json:"-"` // SHA256 hash
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"` // Для одноразовости
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (t *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

func (t *PasswordResetToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// ForgotPasswordRequest - запрос на восстановление пароля
type ForgotPasswordRequest struct {
	Email          string `json:"email"`
	RecaptchaToken string `json:"recaptchaToken"`
}

// ResetPasswordRequest - запрос на сброс пароля
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}
