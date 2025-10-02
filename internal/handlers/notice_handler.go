package handlers

import (
	"context"
	"net/http"
	"time"

	"bms-backend/internal/models"
	"bms-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NoticeHandler struct {
	db *mongo.Database
}

func NewNoticeHandler(db *mongo.Database) *NoticeHandler {
	return &NoticeHandler{db: db}
}

func (h *NoticeHandler) GetNotices(c *gin.Context) {
	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["is_active"] = true

	collection := h.db.Collection("notices")
	cursor, err := collection.Find(context.Background(), societyFilter, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notices"})
		return
	}
	defer cursor.Close(context.Background())

	var notices []models.Notice
	if err = cursor.All(context.Background(), &notices); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode notices"})
		return
	}

	if notices == nil {
		notices = []models.Notice{}
	}

	c.JSON(http.StatusOK, notices)
}

func (h *NoticeHandler) CreateNotice(c *gin.Context) {
	var notice models.Notice
	if err := c.ShouldBindJSON(&notice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authorID, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))
	societyCode := c.GetString("society_code")

	// Get society ID
	societyCollection := h.db.Collection("societies")
	var society models.Society
	err := societyCollection.FindOne(context.Background(), bson.M{"code": societyCode}).Decode(&society)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Society not found"})
		return
	}

	notice.ID = primitive.NewObjectID()
	notice.AuthorID = authorID
	notice.SocietyID = society.ID
	notice.SocietyCode = societyCode
	notice.IsActive = true
	notice.CreatedAt = time.Now()

	collection := h.db.Collection("notices")
	_, err = collection.InsertOne(context.Background(), notice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notice"})
		return
	}

	c.JSON(http.StatusCreated, notice)
}

func (h *NoticeHandler) UpdateNotice(c *gin.Context) {
	noticeID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(noticeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notice ID"})
		return
	}

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["_id"] = objID

	collection := h.db.Collection("notices")
	result, err := collection.UpdateOne(context.Background(), societyFilter, bson.M{"$set": updates})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notice"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notice not found in your society"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notice updated successfully"})
}

func (h *NoticeHandler) DeleteNotice(c *gin.Context) {
	noticeID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(noticeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notice ID"})
		return
	}

	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["_id"] = objID

	update := bson.M{
		"$set": bson.M{
			"is_active": false,
		},
	}

	collection := h.db.Collection("notices")
	result, err := collection.UpdateOne(context.Background(), societyFilter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notice"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notice not found in your society"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notice deleted successfully"})
}