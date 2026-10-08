package account

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewAccountHandler(svc *Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) RegisterRoutes(r *gin.IRouter) {

}

func (h *Handler) HandleStartAuth(c *gin.Context) {
	var authReq StartAuthRequest

	if err := json.NewDecoder(c.Request.Body).Decode(&authReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
		})
	}

	phone, ok := validatePhoneNumber(authReq.PhoneNumber)
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "invalid number format",
		})
	}

}
