package repositories

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Clivet-lug/todo-app/backend/models"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUsersTable() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id         SERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			email      VARCHAR(255) NOT NULL UNIQUE,
			password   VARCHAR(255) NOT NULL,
			role       VARCHAR(50)  NOT NULL DEFAULT 'member',
			created_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Print("Error creating users table:", err)
		return err
	}
	fmt.Println("✅ Users table ready")
	return nil
}

func (r *userRepository) CreateUser(name, email, hashedPassword, role string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, role, created_at
	`, name, email, hashedPassword, role).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, name, email, password, role, created_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, name, email, role, created_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) ListMembers() ([]*models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, name, email, role, created_at
		FROM users WHERE role = 'member' ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("error listing members: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, &u)
	}
	return users, nil
}