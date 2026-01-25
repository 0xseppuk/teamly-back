package models

import (
	"encoding/json"
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

type MessageJSON struct {
	ID             uuid.UUID     `json:"id"`
	ConversationID uuid.UUID     `json:"conversation_id"`
	SenderID       uuid.UUID     `json:"sender_id"`
	Content        string        `json:"content"`
	IsRead         bool          `json:"is_read"`
	ReadAt         *string       `json:"read_at,omitempty"`
	CreatedAt      string        `json:"created_at"`
	UpdatedAt      string        `json:"updated_at"`
	Sender         *User         `json:"sender,omitempty"`
	Conversation   *Conversation `json:"conversation,omitempty"`
}

func (m Message) MarshalJSON() ([]byte, error) {
	formatTime := func(t time.Time) string {
		if t.IsZero() {
			// Если дата не установлена, используем текущее время
			return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		}
		return t.UTC().Format("2006-01-02T15:04:05.000Z")
	}

	msgJSON := MessageJSON{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Content:        m.Content,
		IsRead:         m.IsRead,
		CreatedAt:      formatTime(m.CreatedAt),
		UpdatedAt:      formatTime(m.UpdatedAt),
		Sender:         m.Sender,
		Conversation:   m.Conversation,
	}

	if m.ReadAt != nil {
		readAtStr := formatTime(*m.ReadAt)
		msgJSON.ReadAt = &readAtStr
	}

	return json.Marshal(msgJSON)
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	// Устанавливаем CreatedAt и UpdatedAt, если они не установлены
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	return nil
}

func (m *Message) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
