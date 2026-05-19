package handlers

import (
	"encoding/json"
	"net/http"

	ifaces "github.com/Clivet-lug/todo-app/backend/interfaces"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/utils"
)

// AuthHandler is a struct that holds its dependency (the auth service).
type AuthHandler struct {
	svc ifaces.AuthService
}

func NewAuthHandler(svc ifaces.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Message: "Invalid request body",
		})
		return
	}

	user, token, err := h.svc.Register(req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "email already registered" {
			status = http.StatusConflict
		}
		utils.SendJSON(w, status, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	utils.SendJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    models.AuthResponse{Token: token, User: *user},
	})
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Message: "Invalid request body",
		})
		return
	}

	user, token, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false, Message: err.Error(),
		})
		return
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    models.AuthResponse{Token: token, User: *user},
	})
}