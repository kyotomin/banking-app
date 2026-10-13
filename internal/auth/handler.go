package auth

import (
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
	auth.POST("/verify", h.VerifyCode)
	auth.POST("/register", h.Register)
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

func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	err := h.service.VerifyCode(
		c.Request.Context(),
		req.ChallengeID,
		req.Code,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "verification failed",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "phone verified",
	})
}

func (h *AuthHandler) Register(c *gin.Context) {

}
