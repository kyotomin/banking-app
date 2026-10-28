package transfers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransfersHandler struct {
	service *TransfersService
}

func NewTransfersHandler(svc *TransfersService) *TransfersHandler {
	return &TransfersHandler{
		service: svc,
	}
}

func (h *TransfersHandler) RegisterRoutes(r *gin.Engine) {
	transfer := r.Group("/transfers")

	transfer.POST("create", h.HandleCreateTransfer)
}

func (h *TransfersHandler) HandleCreateTransfer(c *gin.Context) {
	var req TransferInfo
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parsing JSON data",
		})

		return
	}
	req.IKey = c.GetHeader("Idempotency-Key")

	transferID, err := h.service.IssueTransfer(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAmount),
			errors.Is(err, ErrInvalidRecipient):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, ErrAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "account not found",
			})

		case errors.Is(err, ErrInsufficientFunds):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "insufficient funds",
			})

		case errors.Is(err, ErrCurrencyMismatch):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "currency mismatch",
			})

		case errors.Is(err, ErrIdempotencyConflict):
			c.JSON(http.StatusConflict, gin.H{
				"error": "idempotency key reused with different parameters",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transfer_id": transferID,
	})
}
