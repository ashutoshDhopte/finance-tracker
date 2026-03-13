package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	gmailsvc "github.com/ash/finance-tracker/backend/internal/services/gmail"
)

type SyncHandler struct {
	gmailSvc *gmailsvc.Service
}

func NewSyncHandler(gmailSvc *gmailsvc.Service) *SyncHandler {
	return &SyncHandler{gmailSvc: gmailSvc}
}

func (h *SyncHandler) SyncGmail(c *gin.Context) {
	if h.gmailSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Gmail is not configured. Place credentials.json and token.json in the backend directory and restart.",
		})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 {
		days = 1
	}
	if days > 365 {
		days = 365
	}

	ctx := context.Background()
	imported, skipped, failed := h.gmailSvc.SyncDays(ctx, days)

	c.JSON(http.StatusOK, gin.H{
		"message":  "sync complete",
		"days":     days,
		"imported": imported,
		"skipped":  skipped,
		"failed":   failed,
	})
}
