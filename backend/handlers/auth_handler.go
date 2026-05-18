package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Clivet-lug/todo-app/backend/auth"
	"github.com/Clivet-lug/todo-app/backend/databases"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/repositories"
	"github.com/Clivet-lug/todo-app/backend/utils"
	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user account.
// POST /auth/register
// Body: { "name": "Alice", "email": "alice@example.com", "password": "secret", "role": "member" }
//
// Role defaults to "member". Only supply "admin" during initial setup or from
// a trusted admin-only onboarding flow — in production you'd want to gate this.
func Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validation
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Name == "" || req.Email == "" || req.Password == "" {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Name, email, and password are required",
		})
		return
	}
	if len(req.Password) < 6 {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Password must be at least 6 characters",
		})
		return
	}

	role := req.Role
	if role != "admin" && role != "member" {
		role = "member"
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error processing request",
		})
		return
	}

	// Persist user
	db := databases.ConnectDB()
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)
	user, err := userRepo.CreateUser(req.Name, req.Email, string(hash), role)
	if err != nil {
		// Postgres unique_violation code is "23505"
		// refactor ()
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			utils.SendJSON(w, http.StatusConflict, models.APIResponse{
				Success: false,
				Message: "Email already registered",
			})
			return
		}
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error creating user",
		})
		return
	}

	// Issue token 
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error generating token",
		})
		return
	}

	utils.SendJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data: models.AuthResponse{
			Token: token,
			User:  *user,
		},
	})
}

// Login authenticates a user and returns a JWT.
// POST /auth/login
// Body: { "email": "alice@example.com", "password": "secret" }
func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Email and password are required",
		})
		return
	}

	db := databases.ConnectDB()
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)
	user, err := userRepo.FindByEmail(req.Email)
	if err != nil {
		// Don't reveal whether the email exists — generic message
		utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Error generating token",
		})
		return
	}

	utils.SendJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data: models.AuthResponse{
			Token: token,
			User:  *user,
		},
	})
}