package handlers

import (
	"context"
	"net/http"

	"bms-backend/internal/models"
	"bms-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserHandler struct {
	db *mongo.Database
}

func NewUserHandler(db *mongo.Database) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetResidents(c *gin.Context) {
	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["role"] = bson.M{"$in": []string{"resident", "secretary"}}
	societyFilter["is_active"] = true

	collection := h.db.Collection("users")
	cursor, err := collection.Find(context.Background(), societyFilter, options.Find().SetSort(bson.M{"name": 1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch residents"})
		return
	}
	defer cursor.Close(context.Background())

	var users []models.User
	if err = cursor.All(context.Background(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode residents"})
		return
	}

	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}

	if users == nil {
		users = []models.User{}
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) GetStats(c *gin.Context) {
	societyFilter := middleware.GetSocietyFilter(c)
	ctx := context.Background()

	// Count total users in society
	usersCollection := h.db.Collection("users")
	userFilter := societyFilter
	userFilter["is_active"] = true
	totalUsers, _ := usersCollection.CountDocuments(ctx, userFilter)

	// Count residents in society
	residentFilter := societyFilter
	residentFilter["role"] = "resident"
	residentFilter["is_active"] = true
	totalResidents, _ := usersCollection.CountDocuments(ctx, residentFilter)

	// Count pending visitors in society
	visitorsCollection := h.db.Collection("visitors")
	visitorFilter := societyFilter
	visitorFilter["status"] = "pending"
	pendingVisitors, _ := visitorsCollection.CountDocuments(ctx, visitorFilter)

	// Count overdue maintenance in society
	maintenanceCollection := h.db.Collection("maintenance")
	maintenanceFilter := societyFilter
	maintenanceFilter["status"] = "overdue"
	overdueMaintenance, _ := maintenanceCollection.CountDocuments(ctx, maintenanceFilter)

	stats := map[string]interface{}{
		"total_users":        totalUsers,
		"total_residents":    totalResidents,
		"pending_visitors":   pendingVisitors,
		"overdue_maintenance": overdueMaintenance,
		"society_code":       c.GetString("society_code"),
	}

	c.JSON(http.StatusOK, stats)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["_id"] = objID

	collection := h.db.Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), societyFilter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in your society"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetPendingVisitors(c *gin.Context) {
	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["status"] = "pending"

	collection := h.db.Collection("visitors")
	cursor, err := collection.Find(context.Background(), societyFilter, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending visitors"})
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