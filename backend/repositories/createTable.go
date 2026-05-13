package repositories

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Clivet-lug/todo-app/backend/models"
)

type Connector struct {
	DB *sql.DB
}

func NewConnector(db *sql.DB) *Connector {
	return &Connector{DB: db}
}

// CREATE TABLE
func (c *Connector) CreateTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS todos (
			id          SERIAL PRIMARY KEY,
			title       VARCHAR(255) NOT NULL,
			description TEXT,
			completed   BOOLEAN DEFAULT FALSE,
			priority    VARCHAR(50) DEFAULT 'medium',
			created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := c.DB.Exec(query)
	if err != nil {
		log.Print("Error creating todos table:", err)
		return err
	}
	fmt.Println("✅ Todos table ready")
	return nil
}

func (c *Connector) GetTodos() ([]*models.Todo, error) {
	fmt.Println("🔍 Querying PostgreSQL")
	rows, err := c.DB.Query(`
        SELECT id, title, description, completed, priority, created_at, updated_at
        FROM todos
        ORDER BY created_at DESC
    `)
	if err != nil {
		log.Print("Error fetching todos:", err)
		return nil, err
	}
	defer rows.Close()

	var todos []*models.Todo
	for rows.Next() {
		var todo models.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.Priority,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			continue
		}
		todos = append(todos, &todo)
	}
	return todos, nil
}


// GET SINGLE TODO
func (c *Connector) GetTodoByID(id int) (*models.Todo, error) {
	var todo models.Todo

	err := c.DB.QueryRow(`
		SELECT id, title, description, completed, priority, created_at, updated_at
		FROM todos WHERE id = $1
	`, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.Priority,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo with ID %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching todo: %w", err)
	}

	return &todo, nil
}
