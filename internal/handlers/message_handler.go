package handlers

import (
	"time"

	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/duker221/teamly/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetUserConversations returns all conversations for the authenticated user
// Optimized: Preloads only necessary fields, sorts by last activity
func GetUserConversations(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var conversations []models.Conversation

	// Optimized query: Only load what's needed for the list view
	err = database.DB.
		Preload("Participant1", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nickname", "avatar_url")
		}).
		Preload("Participant2", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nickname", "avatar_url")
		}).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			// Load only the last message for preview
			return db.Order("created_at DESC").Limit(1).Select("id", "conversation_id", "sender_id", "content", "created_at", "is_read")
		}).
		Where("participant1_id = ? OR participant2_id = ?", userID, userID).
		Where("is_archived = ?", false).
		Order("last_message_at DESC NULLS LAST").
		Find(&conversations).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch conversations",
		})
	}

	// Calculate unread count for each conversation
	type ConversationResponse struct {
		models.Conversation
		UnreadCount int       `json:"unread_count"`
		OtherUser   *models.User `json:"other_user"`
	}

	response := make([]ConversationResponse, len(conversations))
	for i, conv := range conversations {
		var unreadCount int64
		database.DB.Model(&models.Message{}).
			Where("conversation_id = ? AND sender_id != ? AND is_read = ?", conv.ID, userID, false).
			Count(&unreadCount)

		// Determine who is the "other" user in the conversation
		var otherUser *models.User
		if conv.Participant1ID == userID {
			otherUser = conv.Participant2
		} else {
			otherUser = conv.Participant1
		}

		response[i] = ConversationResponse{
			Conversation: conv,
			UnreadCount:  int(unreadCount),
			OtherUser:    otherUser,
		}
	}

	return c.JSON(response)
}

// GetConversationByID returns a specific conversation with its details
func GetConversationByID(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid conversation ID"})
	}

	var conversation models.Conversation
	err = database.DB.
		Preload("Participant1").
		Preload("Participant2").
		Preload("Response.Application.Game").
		First(&conversation, "id = ?", conversationID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Conversation not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch conversation"})
	}

	// Security: Check if user is a participant
	if conversation.Participant1ID != userID && conversation.Participant2ID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	return c.JSON(conversation)
}

// GetConversationMessages returns messages for a specific conversation
// Supports pagination for performance with large message histories
func GetConversationMessages(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid conversation ID"})
	}

	// Verify user is a participant
	var conversation models.Conversation
	err = database.DB.Select("id", "participant1_id", "participant2_id").
		First(&conversation, "id = ?", conversationID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Conversation not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify conversation"})
	}

	if conversation.Participant1ID != userID && conversation.Participant2ID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Pagination parameters
	limit := c.QueryInt("limit", 50)  // Default 50 messages
	offset := c.QueryInt("offset", 0)

	// Cap limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	var messages []models.Message
	var totalCount int64

	// Get total count for pagination metadata
	database.DB.Model(&models.Message{}).
		Where("conversation_id = ?", conversationID).
		Count(&totalCount)

	// Fetch messages with sender info
	err = database.DB.
		Preload("Sender", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nickname", "avatar_url")
		}).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch messages",
		})
	}

	return c.JSON(fiber.Map{
		"messages": messages,
		"pagination": fiber.Map{
			"total":  totalCount,
			"limit":  limit,
			"offset": offset,
			"has_more": totalCount > int64(offset+limit),
		},
	})
}

// MarkMessagesAsRead marks all messages in a conversation as read
// Optimized: Single UPDATE query instead of updating each message individually
func MarkMessagesAsRead(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid conversation ID"})
	}

	// Verify user is a participant
	var conversation models.Conversation
	err = database.DB.Select("id", "participant1_id", "participant2_id").
		First(&conversation, "id = ?", conversationID).Error

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Conversation not found"})
	}

	if conversation.Participant1ID != userID && conversation.Participant2ID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Bulk update: Mark all unread messages from the other user as read
	now := time.Now()
	result := database.DB.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND is_read = ?", conversationID, userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark messages as read",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"marked_count": result.RowsAffected,
	})
}

// GetUnreadCount returns total unread messages count for the user
func GetUnreadCount(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get all conversation IDs where user is a participant
	var conversationIDs []uuid.UUID
	database.DB.Model(&models.Conversation{}).
		Select("id").
		Where("(participant1_id = ? OR participant2_id = ?) AND is_archived = ?", userID, userID, false).
		Pluck("id", &conversationIDs)

	// Count unread messages across all conversations
	var unreadCount int64
	database.DB.Model(&models.Message{}).
		Where("conversation_id IN ? AND sender_id != ? AND is_read = ?", conversationIDs, userID, false).
		Count(&unreadCount)

	return c.JSON(fiber.Map{
		"unread_count": unreadCount,
	})
}
