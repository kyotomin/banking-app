package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}

}

func (h *DashboardHandler) RegisterRoutes(r *gin.Engine) {
	dashboard := r.Group("/dashboard")

	dashboard.GET("", h.HandleGetDashboard)
}

func (h *DashboardHandler) HandleGetDashboard(c *gin.Context) {
	value, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})

		return
	}

	accountID, ok := value.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid account id",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id": accountID,
	})
}
