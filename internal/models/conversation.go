package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Conversation struct {
	ID         uuid.UUID            `gorm:"primaryKey" json:"id"`
	ResponseID uuid.UUID            `gorm:"not null;uniqueIndex" json:"response_id"`
	Response   *ApplicationResponse `gorm:"foreignKey:ResponseID" json:"response,omitempty"`

	Participant1ID uuid.UUID `gorm:"not null;index:idx_participants" json:"participant1_id"`
	Participant1   *User     `gorm:"foreignKey:Participant1ID" json:"participant1,omitempty"`
	Participant2ID uuid.UUID `gorm:"not null;index:idx_participants" json:"participant2_id"`
	Participant2   *User     `gorm:"foreignKey:Participant2ID" json:"participant2,omitempty"`

	LastMessageAt *time.Time `gorm:"index" json:"last_message_at,omitempty"`
	IsArchived    bool       `gorm:"default:false;index" json:"is_archived"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type Message struct {
	ID             uuid.UUID     `gorm:"primaryKey" json:"id"`
	ConversationID uuid.UUID     `gorm:"not null;index:idx_conversation_created" json:"conversation_id"`
	Conversation   *Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`

	SenderID uuid.UUID `gorm:"not null;index" json:"sender_id"`
	Sender   *User     `gorm:"foreignKey:SenderID" json:"sender,omitempty"`

	Content string `gorm:"type:text;not null" json:"content"`

	IsRead bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt *time.Time `json:"read_at,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_conversation_created" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
