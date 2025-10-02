package handlers

import (
	"context"
	"net/http"
	"time"

	"bms-backend/internal/models"
	"bms-backend/pkg/auth"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db        *mongo.Database
	jwtSecret string
}

func NewAuthHandler(db *mongo.Database, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (h *AuthHandler) ValidateSociety(c *gin.Context) {
	var req models.SocietyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find society by code
	collection := h.db.Collection("societies")
	var society models.Society
	err := collection.FindOne(context.Background(), bson.M{"code": req.Code, "is_active": true}).Decode(&society)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid society code"})
		return
	}

	societyResponse := models.SocietyResponse{
		ID:      society.ID,
		Name:    society.Name,
		Code:    society.Code,
		Address: society.Address,
		City:    society.City,
		State:   society.State,
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"society": societyResponse,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First validate society code
	societyCollection := h.db.Collection("societies")
	var society models.Society
	err := societyCollection.FindOne(context.Background(), bson.M{"code": req.SocietyCode, "is_active": true}).Decode(&society)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid society code"})
		return
	}

	// Find user by email and society code
	collection := h.db.Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{
		"email":        req.Email,
		"society_code": req.SocietyCode,
		"is_active":    true,
	}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token with society info
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, user.SocietyCode, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Remove password from response
	user.Password = ""

	societyResponse := models.SocietyResponse{
		ID:      society.ID,
		Name:    society.Name,
		Code:    society.Code,
		Address: society.Address,
		City:    society.City,
		State:   society.State,
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:   token,
		User:    user,
		Society: societyResponse,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First validate society code
	societyCollection := h.db.Collection("societies")
	var society models.Society
	err := societyCollection.FindOne(context.Background(), bson.M{"code": req.SocietyCode, "is_active": true}).Decode(&society)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid society code"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Email:       req.Email,
		Password:    string(hashedPassword),
		Role:        req.Role,
		Unit:        req.Unit,
		Building:    req.Building,
		Phone:       req.Phone,
		SocietyID:   society.ID,
		SocietyCode: req.SocietyCode,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	collection := h.db.Collection("users")
	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists in this society"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	societyCode := c.GetString("society_code")
	objID, _ := primitive.ObjectIDFromHex(userID)

	collection := h.db.Collection("users")
	var user models.User
	err := collection.FindOne(context.Background(), bson.M{
		"_id":          objID,
		"society_code": societyCode,
	}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusOK, user)
}