package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pavelc4/auriya-todolist-go/internal/http/repository"
	"github.com/pavelc4/auriya-todolist-go/internal/http/service"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	GoogleConfig *oauth2.Config
	UserRepo     *repository.UserRepository
	JWTService   *service.JWTService
}

func NewAuthHandler(googleConfig *oauth2.Config, userRepo *repository.UserRepository, jwtService *service.JWTService) *AuthHandler {
	return &AuthHandler{
		GoogleConfig: googleConfig,
		UserRepo:     userRepo,
		JWTService:   jwtService,
	}
}

// Redirect to login page Google
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := "random"
	url := h.GoogleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Callback dan exchange token from Google
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := h.GoogleConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed"})
		return
	}

	// Colect data profil user from Google API
	client := h.GoogleConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var profile struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user info"})
		return
	}

	if !profile.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
		return
	}

	// Save update user on  database
	user, err := h.UserRepo.GetByProviderUserID(c.Request.Context(), "google", profile.Sub)
	if err != nil {
		user = &repository.User{
			Email:          profile.Email,
			FullName:       profile.Name,
			AvatarURL:      profile.Picture,
			Provider:       "google",
			ProviderUserID: profile.Sub,
		}
		_ = h.UserRepo.Create(c.Request.Context(), user)
	} else {
		_ = h.UserRepo.UpdateLastLogin(c.Request.Context(), user.ID)
	}
	// Create JWT token
	jwtToken, err := h.JWTService.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
		"user":  user,
	})
}
