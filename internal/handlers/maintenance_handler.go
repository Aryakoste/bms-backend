package handlers

import (
	"context"
	"net/http"
	"time"

	"bms-backend/internal/middleware"
	"bms-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MaintenanceHandler struct {
	db *mongo.Database
}

func NewMaintenanceHandler(db *mongo.Database) *MaintenanceHandler {
	return &MaintenanceHandler{db: db}
}

func (h *MaintenanceHandler) GetMaintenanceByID(c *gin.Context) {
	noticeID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(noticeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance ID"})
		return
	}

	collection := h.db.Collection("maintenance")
	var maintenance models.MaintenanceRecord
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&maintenance)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, maintenance)
}

func (h *MaintenanceHandler) GetMaintenanceRecords(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	// societyCode := c.GetString("society_code")

	// Base filter with society
	filter := middleware.GetSocietyFilter(c)

	if userRole == "resident" {
		// Residents can only see their own maintenance records
		objID, _ := primitive.ObjectIDFromHex(userID)
		filter["unit_id"] = objID
	}

	collection := h.db.Collection("maintenance")
	cursor, err := collection.Find(context.Background(), filter, options.Find().SetSort(bson.M{"due_date": -1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch maintenance records"})
		return
	}
	defer cursor.Close(context.Background())

	var records []models.MaintenanceRecord
	if err = cursor.All(context.Background(), &records); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode maintenance records"})
		return
	}

	if records == nil {
		records = []models.MaintenanceRecord{}
	}

	c.JSON(http.StatusOK, records)
}

func (h *MaintenanceHandler) PayMaintenance(c *gin.Context) {
	var req models.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maintenanceID, err := primitive.ObjectIDFromHex(req.MaintenanceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance ID"})
		return
	}

	// Society filter for security
	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["_id"] = maintenanceID

	// Simulate payment processing
	paymentID := primitive.NewObjectID().Hex()
	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"status":      "paid",
			"paid_date":   now,
			"receipt_url": "/receipts/" + paymentID + ".pdf",
		},
	}

	collection := h.db.Collection("maintenance")
	result, err := collection.UpdateOne(context.Background(), societyFilter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance record not found in your society"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Payment processed successfully",
		"payment_id":  paymentID,
		"receipt_url": "/receipts/" + paymentID + ".pdf",
	})
}

func (h *MaintenanceHandler) CreateMaintenanceRecord(c *gin.Context) {
	var record models.MaintenanceRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	societyCode := c.GetString("society_code")

	// Get society ID
	societyCollection := h.db.Collection("societies")
	var society models.Society
	err := societyCollection.FindOne(context.Background(), bson.M{"code": societyCode}).Decode(&society)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Society not found"})
		return
	}

	record.ID = primitive.NewObjectID()
	record.Status = "pending"
	record.SocietyID = society.ID
	record.SocietyCode = societyCode
	record.CreatedAt = time.Now()

	// If no unit_id provided, create a dummy one
	if record.UnitID.IsZero() {
		record.UnitID = primitive.NewObjectID()
	}

	collection := h.db.Collection("maintenance")
	_, err = collection.InsertOne(context.Background(), record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance record"})
		return
	}

	c.JSON(http.StatusCreated, record)
}
