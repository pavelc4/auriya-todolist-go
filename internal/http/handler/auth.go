package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pavelc4/auriya-todolist-go/internal/http/repository"
	"github.com/pavelc4/auriya-todolist-go/internal/http/service"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	GoogleConfig *oauth2.Config
	GitHubConfig *oauth2.Config
	UserRepo     *repository.UserRepository
	JWTService   *service.JWTService
}

func NewAuthHandler(
	googleConfig *oauth2.Config,
	githubConfig *oauth2.Config,
	userRepo *repository.UserRepository,
	jwtService *service.JWTService,
) *AuthHandler {
	return &AuthHandler{
		GoogleConfig: googleConfig,
		GitHubConfig: githubConfig,
		UserRepo:     userRepo,
		JWTService:   jwtService,
	}
}

func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	state := "randomstate"
	url := h.GitHubConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Google Login redirect
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := "randomstate" // ideally use a random string and store in session
	url := h.GoogleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Google callback handler
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

	user, err := h.UserRepo.GetByProviderUserID(c.Request.Context(), "google", profile.Sub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error", "detail": err.Error()})
		return
	}

	if user == nil {
		user = &repository.User{
			Email:          profile.Email,
			FullName:       profile.Name,
			AvatarURL:      profile.Picture,
			Provider:       "google",
			ProviderUserID: profile.Sub,
		}
		err = h.UserRepo.Create(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user create failed", "detail": err.Error()})
			return
		}
	} else {
		_ = h.UserRepo.UpdateLastLogin(c.Request.Context(), user.ID)
	}

	tokenString, err := h.JWTService.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := h.GitHubConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed"})
		return
	}

	client := h.GitHubConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var profile struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user info"})
		return
	}

	// Get user (by provider + provider user id)
	user, err := h.UserRepo.GetByProviderUserID(c.Request.Context(), "github", strconv.Itoa(profile.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error", "detail": err.Error()})
		return
	}
	if user == nil {
		user = &repository.User{
			Email:          profile.Email,
			FullName:       profile.Name,
			AvatarURL:      profile.AvatarURL,
			Provider:       "github",
			ProviderUserID: strconv.Itoa(profile.ID),
		}
		err = h.UserRepo.Create(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user create failed", "detail": err.Error()})
			return
		}
	} else {
		_ = h.UserRepo.UpdateLastLogin(c.Request.Context(), user.ID)
	}

	// JWT sama seperti Google
	tokenString, err := h.JWTService.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

// Manual Register handler
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash_failed"})
		return
	}

	user := &repository.User{
		Email:          req.Email,
		FullName:       req.FullName,
		Age:            req.Age,
		Password:       string(hashed),
		Provider:       "local",
		ProviderUserID: req.Email,
	}

	existingUser, err := h.UserRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error", "detail": err.Error()})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user_exists"})
		return
	}

	err = h.UserRepo.Create(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_create_failed", "detail": err.Error()})
		return
	}

	tokenString, err := h.JWTService.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "register success",
		"token":   tokenString,
		"user":    user,
	})
}

// Manual login handler
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}

	user, err := h.UserRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error", "detail": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}

	if user.Provider != "local" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "use_oauth_login"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}

	_ = h.UserRepo.UpdateLastLogin(c.Request.Context(), user.ID)

	tokenString, err := h.JWTService.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}
