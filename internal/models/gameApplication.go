package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Platform string

const (
	PlatformPC             Platform = "pc"
	PlatformPlayStation    Platform = "playstation"
	PlatformXbox           Platform = "xbox"
	PlatformNintendoSwitch Platform = "nintendo_switch"
	PlatformMobile         Platform = "mobile"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusAccepted Status = "accepted"
	StatusRejected Status = "rejected"
)

type GameApplication struct {
	ID     uuid.UUID `gorm:"primaryKey" json:"id"`
	UserId uuid.UUID `gorm:"not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserId" json:"user"`
	GameId uuid.UUID `gorm:"not null" json:"game_id"`
	Game   Game      `gorm:"foreignKey:GameId" json:"game"`

	Title       string `gorm:"not null" json:"title"`
	Description string `gorm:"not null" json:"description"`

	MaxPlayers      int `gorm:"not null" json:"max_players"`
	MinPlayers      int `gorm:"not null" json:"min_players"`
	AcceptedPlayers int `gorm:"not null" json:"accepted_players"`

	PrimeTimeStart time.Time `gorm:"not null" json:"prime_time_start"`
	PrimeTimeEnd   time.Time `gorm:"not null" json:"prime_time_end"`

	IsActive      bool `gorm:"default:true;index" json:"is_active"`
	IsFull        bool `gorm:"default:false" json:"is_full"`
	WithVoiceChat bool `gorm:"default:false" json:"with_voice_chat"`

	Platform  Platform  `gorm:"default:pc" json:"platform"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type ApplicationResponse struct {
	ID            uuid.UUID        `gorm:"primaryKey" json:"id"`
	ApplicationID uuid.UUID        `gorm:"not null;index" json:"application_id"`
	Application   *GameApplication `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	UserID        uuid.UUID        `gorm:"not null;index" json:"user_id"` // Кто откликнулся
	User          *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`

	Status        Status           `gorm:"default:'pending';index" json:"status"`

	CreatedAt     time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time        `gorm:"autoUpdateTime" json:"updated_at"`

	// Связь 1:1 с Conversation
	Conversation  *Conversation    `gorm:"foreignKey:ResponseID" json:"conversation,omitempty"`
}

func (ga *GameApplication) BeforeCreate(tx *gorm.DB) error {
	if ga.ID == uuid.Nil {
		ga.ID = uuid.New()
	}
	if ga.AcceptedPlayers == 0 {
		ga.AcceptedPlayers = 0
	}
	return nil
}

func (ar *ApplicationResponse) BeforeCreate(tx *gorm.DB) error {
	if ar.ID == uuid.Nil {
		ar.ID = uuid.New()
	}
	return nil
}
