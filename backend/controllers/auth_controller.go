package controllers

import (
	"net/http"
	"strings"

	"live-polling/backend/middleware"
	"live-polling/backend/models"
	"live-polling/backend/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	mongoRepo *repository.MongoRepo
	jwtSecret string
}

func NewAuthController(mongoRepo *repository.MongoRepo, jwtSecret string) *AuthController {
	return &AuthController{
		mongoRepo: mongoRepo,
		jwtSecret: jwtSecret,
	}
}

func (ac *AuthController) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if len(req.Name) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be at least 2 characters long"})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters long"})
		return
	}

	// Check if user already exists
	existingUser, err := ac.mongoRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password"})
		return
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := ac.mongoRepo.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := middleware.GenerateToken(user.ID.Hex(), user.Email, ac.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": models.UserResponse{
			ID:        user.ID.Hex(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			Token:     token,
		},
	})
}

func (ac *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please enter both email/username and password"})
		return
	}

	identifier := strings.TrimSpace(req.Email)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Identifier)
	}

	if identifier == "" || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please enter both email/username and password"})
		return
	}

	user, err := ac.mongoRepo.GetUserByEmailOrName(c.Request.Context(), identifier)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email/username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email/username or password"})
		return
	}

	token, err := middleware.GenerateToken(user.ID.Hex(), user.Email, ac.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": models.UserResponse{
			ID:        user.ID.Hex(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			Token:     token,
		},
	})
}

func (ac *AuthController) Me(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID := userIDVal.(primitive.ObjectID)
	user, err := ac.mongoRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": models.UserResponse{
			ID:        user.ID.Hex(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}
