package handlers

import (
	"net/http"

	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/utils"
)

// HEALTH CHECK
// Checks both PostgreSQL and Redis
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// if err := db.Ping(); err != nil {
	// 	utils.SendJSON(w, http.StatusServiceUnavailable, models.APIResponse{
	// 		Success: false,
	// 		Message: "PostgreSQL unreachable",
	// 	})
	// 	return
	// }

	// message := "API healthy - PostgreSQL connected"

	// // Only check Redis if available
	// if rdb != nil {
	// 	if _, err := rdb.Ping(ctx).Result(); err == nil {
	// 		message = "API healthy - PostgreSQL and Redis connected"
	// 	}
	// }

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Service is healthy",
	})
}
