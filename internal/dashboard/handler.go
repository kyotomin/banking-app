package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service *DashboardService
}

func NewDashboardHandler(svc *DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service: svc,
	}
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
		return
	}

	dashboard, err := h.service.GetDashboard(
		c.Request.Context(),
		accountID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
