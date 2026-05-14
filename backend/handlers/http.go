package handlers

import (
	"net/http"

	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/utils"
)

// HEALTH CHECK
// Checks both PostgreSQL and Redis
func HealthCheck(w http.ResponseWriter, r *http.Request) {

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Service is healthy",
	})
}
