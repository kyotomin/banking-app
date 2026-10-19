package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(svc *AuthService) *AuthHandler {
	return &AuthHandler{
		service: svc,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")

	auth.POST("/start", h.HandleStartAuth)
	auth.POST("/register", h.Register)
	auth.POST("/confirm", h.HandleConfirmCode)
	auth.POST("/login", h.Login)
}

func (h *AuthHandler) HandleStartAuth(c *gin.Context) {
	var authReq StartAuthRequest

	if err := c.ShouldBindJSON(&authReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
		})
		return
	}

	challenge, err := h.service.StartAuth(
		c.Request.Context(),
		authReq.PhoneNumber,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, "unable to start challenge")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"challenge_id": challenge.ID,
	})
}

func (h *AuthHandler) HandleConfirmCode(c *gin.Context) {
	var req VerifyCodeRequest

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	result, err := h.service.ConfirmCode(
		c.Request.Context(),
		req.ChallengeID,
		req.Code,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidOTP):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
		case errors.Is(err, ErrTooManyAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterAccountRequest

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad registration request",
		})
		return
	}

	tokens, err := h.service.Register(
		c.Request.Context(),
		req.Token,
		req.Owner,
		req.Pin,
		req.Currency,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "register failed",
		})
		return
	}

	c.JSON(http.StatusCreated, tokens)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginAccountRequest

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	tokens, err := h.service.Login(
		c.Request.Context(),
		req.PhoneNumber,
		req.Pin,
	)

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) ||
			errors.Is(err, ErrAccountNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid phone or PIN",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}
