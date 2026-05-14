package repositories

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Clivet-lug/todo-app/backend/models"
)

// UserRepository handles all DB operations for users.
type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUsersTable creates the users table if it doesn't exist.
// Call this once at startup, before CreateTodosTable (todos references users).
func (u *UserRepository) CreateUsersTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id         SERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			email      VARCHAR(255) NOT NULL UNIQUE,
			password   VARCHAR(255) NOT NULL,  -- bcrypt hash
			role       VARCHAR(50)  NOT NULL DEFAULT 'member',
			created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := u.DB.Exec(query)
	if err != nil {
		log.Print("Error creating users table:", err)
		return err
	}
	fmt.Println("✅ Users table ready")
	return nil
}

// CreateUser inserts a new user and returns the created record.
// The password field should already be a bcrypt hash before calling this.
func (u *UserRepository) CreateUser(name, email, hashedPassword, role string) (*models.User, error) {
	var user models.User
	err := u.DB.QueryRow(`
		INSERT INTO users (name, email, password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, role, created_at
	`, name, email, hashedPassword, role).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	return &user, nil
}

// FindByEmail looks up a user by email address.
// Returns sql.ErrNoRows (wrapped) if not found.
func (u *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := u.DB.QueryRow(`
		SELECT id, name, email, password, role, created_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	return &user, nil
}

// FindByID looks up a user by primary key.
func (u *UserRepository) FindByID(id int) (*models.User, error) {
	var user models.User
	err := u.DB.QueryRow(`
		SELECT id, name, email, role, created_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	return &user, nil
}

// ListMembers returns all users with the "member" role.
// Useful for the admin assignment UI.
func (u *UserRepository) ListMembers() ([]*models.User, error) {
	rows, err := u.DB.Query(`
		SELECT id, name, email, role, created_at
		FROM users WHERE role = 'member'
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("error listing members: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt); err != nil {
			continue
		}
		users = append(users, &user)
	}
	return users, nil
}
