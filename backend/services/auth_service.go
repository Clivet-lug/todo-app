package services

import (
	"fmt"
	"strings"

	"github.com/Clivet-lug/todo-app/backend/auth"
	ifaces "github.com/Clivet-lug/todo-app/backend/interfaces"
	"github.com/Clivet-lug/todo-app/backend/models"
	"golang.org/x/crypto/bcrypt"
)

// authService holds business logic for registration and login.
// It depends on the UserRepository INTERFACE — not the concrete type.
type authService struct {
	userRepo ifaces.UserRepository
}

func NewAuthService(userRepo ifaces.UserRepository) ifaces.AuthService {
	return &authService{userRepo: userRepo}
}

// Register validates input, checks for duplicate email,
// hashes the password, persists the user, and issues a JWT.
func (s *authService) Register(name, email, password, role string) (*models.User, string, error) {
	// ── Normalise input ───────────────────────────────────────────────────────
	name     = strings.TrimSpace(name)
	email    = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)

	if name == "" || email == "" || password == "" {
		return nil, "", fmt.Errorf("name, email and password are required")
	}
	if len(password) < 6 {
		return nil, "", fmt.Errorf("password must be at least 6 characters")
	}
	if role != "admin" && role != "member" {
		role = "member"
	}

	// Business rule: email must be unique
	// This check lives in the SERVICE, not the handler or repository.
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, "", fmt.Errorf("email already registered")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("error processing password")
	}

	// Persist
	user, err := s.userRepo.CreateUser(name, email, string(hash), role)
	if err != nil {
		return nil, "", fmt.Errorf("error creating user")
	}

	// Issue token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("error generating token")
	}

	return user, token, nil
}

// Login verifies credentials and issues a JWT.
func (s *authService) Login(email, password string) (*models.User, string, error) {
	email    = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return nil, "", fmt.Errorf("email and password are required")
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Generic message — don't reveal whether the email exists
		return nil, "", fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("error generating token")
	}

	return user, token, nil
}