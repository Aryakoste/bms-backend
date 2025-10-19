package handlers

import (
	"context"
	"net/http"
	"time"

	"bms-backend/internal/middleware"
	"bms-backend/internal/models"
	"bms-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VisitorHandler struct {
	db *mongo.Database
}

func NewVisitorHandler(db *mongo.Database) *VisitorHandler {
	return &VisitorHandler{db: db}
}

func (h *VisitorHandler) GetVisitors(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	// societyCode := c.GetString("society_code")

	// Base filter with society
	filter := middleware.GetSocietyFilter(c)

	if userRole == "resident" {
		// Residents can only see their own visitors
		hostID, _ := primitive.ObjectIDFromHex(userID)
		filter["host_id"] = hostID
	}

	collection := h.db.Collection("visitors")
	cursor, err := collection.Find(context.Background(), filter, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch visitors"})
		return
	}
	defer cursor.Close(context.Background())

	var visitors []models.Visitor
	if err = cursor.All(context.Background(), &visitors); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode visitors"})
		return
	}

	if visitors == nil {
		visitors = []models.Visitor{}
	}

	c.JSON(http.StatusOK, visitors)
}

func (h *VisitorHandler) CreateVisitor(c *gin.Context) {
	var visitor models.Visitor
	if err := c.ShouldBindJSON(&visitor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	societyCode := c.GetString("society_code")
	hostID, _ := primitive.ObjectIDFromHex(userID)

	// Get society ID
	societyCollection := h.db.Collection("societies")
	var society models.Society
	err := societyCollection.FindOne(context.Background(), bson.M{"code": societyCode}).Decode(&society)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Society not found"})
		return
	}

	visitor.ID = primitive.NewObjectID()
	visitor.HostID = hostID
	visitor.Status = "pending"
	visitor.QRCode = utils.GenerateQRCode(visitor.ID.Hex(), societyCode)
	visitor.SocietyID = society.ID
	visitor.SocietyCode = societyCode
	visitor.CreatedAt = time.Now()
	visitor.UpdatedAt = time.Now()

	// Parse expected time if provided
	if visitor.ExpectedTime.IsZero() {
		visitor.ExpectedTime = time.Now()
	}

	collection := h.db.Collection("visitors")
	_, err = collection.InsertOne(context.Background(), visitor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create visitor"})
		return
	}

	c.JSON(http.StatusCreated, visitor)
}

func (h *VisitorHandler) ApproveVisitor(c *gin.Context) {
	visitorID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(visitorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visitor ID"})
		return
	}

	var req models.VisitorApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approvedBy, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))
	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["_id"] = objID

	update := bson.M{
		"$set": bson.M{
			"status":      req.Status,
			"approved_by": approvedBy,
			"updated_at":  time.Now(),
		},
	}

	collection := h.db.Collection("visitors")
	result, err := collection.UpdateOne(context.Background(), societyFilter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update visitor"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Visitor not found in your society"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visitor " + req.Status + " successfully"})
}

func (h *VisitorHandler) GetVisitorByQR(c *gin.Context) {
	qrCode := c.Param("qrcode")

	collection := h.db.Collection("visitors")
	var visitor models.Visitor
	err := collection.FindOne(context.Background(), bson.M{"qr_code": qrCode}).Decode(&visitor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Visitor not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, visitor)
}

func (h *VisitorHandler) CheckInVisitor(c *gin.Context) {
	visitorID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(visitorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visitor ID"})
		return
	}

	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["_id"] = objID

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"actual_arrival": now,
			"updated_at":     now,
		},
	}

	collection := h.db.Collection("visitors")
	result, err := collection.UpdateOne(context.Background(), societyFilter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in visitor"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Visitor not found in your society"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visitor checked in successfully"})
}

func (h *VisitorHandler) CheckOutVisitor(c *gin.Context) {
	visitorID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(visitorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visitor ID"})
		return
	}

	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["_id"] = objID

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"actual_departure": now,
			"status":           "completed",
			"updated_at":       now,
		},
	}

	collection := h.db.Collection("visitors")
	result, err := collection.UpdateOne(context.Background(), societyFilter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check out visitor"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Visitor not found in your society"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visitor checked out successfully"})
}

func (h *VisitorHandler) GetVisitorByID(c *gin.Context) {
	visitorID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(visitorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visitor ID"})
		return
	}

	collection := h.db.Collection("visitors")
	var visitor models.Visitor
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&visitor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Visitor not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, visitor)
}
