package auth

import (
	"net/http"

	"github.com/baochammm/mangahub/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo *AuthRepository
}

func NewAuthHandler(repo *AuthRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

type SignupInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	err := h.repo.CreateUser(input.Username, string(hashed))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "signup successful"})
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := h.repo.FindByUsername(input.Username)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	tokenString, err := utils.GenerateJWT(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.SetCookie(
		"jwt",
		tokenString,
		3600*24,
		"/",
		"",
		false,
		true,
	) //for browser clients

	c.Header("Authorization", "Bearer "+tokenString) //for CLI clients

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":       user.UserID,
			"username": user.Username,
		},
		"token": tokenString,
	})

}

func (h *AuthHandler) Logout(c *gin.Context) {
	jwtCookie, err := c.Cookie("jwt")
	if err != nil || jwtCookie == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "You are not logged in.",
		})
		return
	}

	c.SetCookie("jwt", "", -1, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
