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

type AmenityHandler struct {
	db *mongo.Database
}

func NewAmenityHandler(db *mongo.Database) *AmenityHandler {
	return &AmenityHandler{db: db}
}

func (h *AmenityHandler) GetAmenities(c *gin.Context) {
	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["is_active"] = true

	collection := h.db.Collection("amenities")
	cursor, err := collection.Find(context.Background(), societyFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch amenities"})
		return
	}
	defer cursor.Close(context.Background())

	var amenities []models.Amenity
	if err = cursor.All(context.Background(), &amenities); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode amenities"})
		return
	}

	if amenities == nil {
		amenities = []models.Amenity{}
	}

	c.JSON(http.StatusOK, amenities)
}

func (h *AmenityHandler) BookAmenity(c *gin.Context) {
	var booking models.AmenityBooking
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))
	societyCode := c.GetString("society_code")

	// Get society ID
	societyCollection := h.db.Collection("societies")
	var society models.Society
	err := societyCollection.FindOne(context.Background(), bson.M{"code": societyCode}).Decode(&society)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Society not found"})
		return
	}

	// Get amenity details (with society check)
	amenityID := booking.AmenityID
	amenityCollection := h.db.Collection("amenities")
	var amenity models.Amenity
	err = amenityCollection.FindOne(context.Background(), bson.M{
		"_id":          amenityID,
		"society_code": societyCode,
	}).Decode(&amenity)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Amenity not found in your society"})
		return
	}

	// Check if time slot is available in the same society
	bookingCollection := h.db.Collection("amenity_bookings")
	existingBooking := bookingCollection.FindOne(context.Background(), bson.M{
		"amenity_id":   amenityID,
		"date":         booking.Date,
		"time_slot":    booking.TimeSlot,
		"society_code": societyCode,
		"status":       bson.M{"$ne": "cancelled"},
	})

	if existingBooking.Err() == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Time slot already booked"})
		return
	}

	booking.ID = primitive.NewObjectID()
	booking.UserID = userID
	booking.AmenityName = amenity.Name
	booking.Status = "confirmed"
	booking.TotalAmount = amenity.BookingFee
	booking.PaymentID = primitive.NewObjectID().Hex()
	booking.SocietyID = society.ID
	booking.SocietyCode = societyCode
	booking.CreatedAt = time.Now()

	_, err = bookingCollection.InsertOne(context.Background(), booking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

func (h *AmenityHandler) GetBookings(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")

	// Base filter with society
	filter := middleware.GetSocietyFilter(c)

	if userRole == "resident" {
		objID, _ := primitive.ObjectIDFromHex(userID)
		filter["user_id"] = objID
	}

	collection := h.db.Collection("amenity_bookings")
	cursor, err := collection.Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}
	defer cursor.Close(context.Background())

	var bookings []models.AmenityBooking
	if err = cursor.All(context.Background(), &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode bookings"})
		return
	}

	if bookings == nil {
		bookings = []models.AmenityBooking{}
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *AmenityHandler) CancelBooking(c *gin.Context) {
	bookingID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	societyFilter := middleware.GetSocietyFilter(c)
	societyFilter["_id"] = objID

	update := bson.M{
		"$set": bson.M{
			"status": "cancelled",
		},
	}

	collection := h.db.Collection("amenity_bookings")
	result, err := collection.UpdateOne(context.Background(), societyFilter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found in your society"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}