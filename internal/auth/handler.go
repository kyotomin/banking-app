package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *AuthService
}

func NewAccountAuthHandler(svc *AuthService) *AuthHandler {
	return &AuthHandler{
		service: svc,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.IRouter) {

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
