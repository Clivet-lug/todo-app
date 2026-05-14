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

// CreateTable creates the todos table with the full schema including
// workflow status and assignment columns.
// It also runs MigrateTable() to safely add new columns to an existing table.
func (c *Connector) CreateTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS todos (
			id          SERIAL PRIMARY KEY,
			title       VARCHAR(255) NOT NULL,
			description TEXT,
			completed   BOOLEAN      DEFAULT FALSE,
			priority    VARCHAR(50)  DEFAULT 'medium',
			status      VARCHAR(50)  DEFAULT 'todo',
			assigned_to INT          REFERENCES users(id) ON DELETE SET NULL,
			assigned_by INT          REFERENCES users(id) ON DELETE SET NULL,
			created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
			updated_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := c.DB.Exec(query)
	if err != nil {
		log.Print("Error creating todos table:", err)
		return err
	}
	fmt.Println("✅ Todos table ready")

	// Run migration in case the table already existed without the new columns
	return c.MigrateTable()
}

// MigrateTable adds new columns to an existing todos table gracefully.
// Each ALTER is wrapped in a DO block so it's a no-op if the column exists.
func (c *Connector) MigrateTable() error {
	migrations := []string{
		`DO $$ BEGIN
			ALTER TABLE todos ADD COLUMN status VARCHAR(50) DEFAULT 'todo';
		EXCEPTION WHEN duplicate_column THEN NULL; END $$`,

		`DO $$ BEGIN
			ALTER TABLE todos ADD COLUMN assigned_to INT REFERENCES users(id) ON DELETE SET NULL;
		EXCEPTION WHEN duplicate_column THEN NULL; END $$`,

		`DO $$ BEGIN
			ALTER TABLE todos ADD COLUMN assigned_by INT REFERENCES users(id) ON DELETE SET NULL;
		EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
	}

	for _, m := range migrations {
		if _, err := c.DB.Exec(m); err != nil {
			log.Print("Migration error:", err)
			return err
		}
	}
	fmt.Println("✅ Todos table migrations applied")
	return nil
}

// ─── SCANNING HELPER ─────────────────────────────────────────────────────────

// scanTodo reads a row that SELECTs all todo columns plus an optional
// LEFT JOIN on users (assignee name + email). Pass nil for rows without
// the JOIN columns.
func scanTodoRow(row interface {
	Scan(...interface{}) error
}, withAssignee bool) (*models.Todo, error) {
	var todo models.Todo
	var assignedTo, assignedBy sql.NullInt64

	if withAssignee {
		var assigneeName, assigneeEmail sql.NullString
		err := row.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.Priority,
			&todo.Status,
			&assignedTo,
			&assignedBy,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&assigneeName,
			&assigneeEmail,
		)
		if err != nil {
			return nil, err
		}
		if assignedTo.Valid {
			v := int(assignedTo.Int64)
			todo.AssignedTo = &v
		}
		if assignedBy.Valid {
			v := int(assignedBy.Int64)
			todo.AssignedBy = &v
		}
		if assigneeName.Valid {
			todo.Assignee = &models.UserSummary{
				ID:    int(assignedTo.Int64),
				Name:  assigneeName.String,
				Email: assigneeEmail.String,
			}
		}
		return &todo, nil
	}

	err := row.Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.Priority,
		&todo.Status,
		&assignedTo,
		&assignedBy,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if assignedTo.Valid {
		v := int(assignedTo.Int64)
		todo.AssignedTo = &v
	}
	if assignedBy.Valid {
		v := int(assignedBy.Int64)
		todo.AssignedBy = &v
	}
	return &todo, nil
}

// ─── QUERIES ──────────────────────────────────────────────────────────────────

// GetTodos returns all todos with assignee info joined.
// Pass userID = 0 and isAdmin = true to get all todos (admin view).
// Pass a real userID and isAdmin = false to get only that user's assigned todos.
func (c *Connector) GetTodos(userID int, isAdmin bool) ([]*models.Todo, error) {
	fmt.Println("🔍 Querying PostgreSQL")

	baseQuery := `
		SELECT t.id, t.title, t.description, t.completed, t.priority,
		       t.status, t.assigned_to, t.assigned_by,
		       t.created_at, t.updated_at,
		       u.name, u.email
		FROM todos t
		LEFT JOIN users u ON u.id = t.assigned_to
	`

	var rows *sql.Rows
	var err error

	if isAdmin {
		rows, err = c.DB.Query(baseQuery + ` ORDER BY t.created_at DESC`)
	} else {
		rows, err = c.DB.Query(
			baseQuery+` WHERE t.assigned_to = $1 ORDER BY t.created_at DESC`,
			userID,
		)
	}

	if err != nil {
		log.Print("Error fetching todos:", err)
		return nil, err
	}
	defer rows.Close()

	var todos []*models.Todo
	for rows.Next() {
		todo, err := scanTodoRow(rows, true)
		if err != nil {
			continue
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

// GetTodoByID returns a single todo with assignee info.
func (c *Connector) GetTodoByID(id int) (*models.Todo, error) {
	row := c.DB.QueryRow(`
		SELECT t.id, t.title, t.description, t.completed, t.priority,
		       t.status, t.assigned_to, t.assigned_by,
		       t.created_at, t.updated_at,
		       u.name, u.email
		FROM todos t
		LEFT JOIN users u ON u.id = t.assigned_to
		WHERE t.id = $1
	`, id)

	todo, err := scanTodoRow(row, true)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo with ID %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching todo: %w", err)
	}
	return todo, nil
}

// AssignTodo sets the assigned_to and assigned_by fields (admin only operation).
func (c *Connector) AssignTodo(todoID, assignedToUserID, assignedByUserID int) (*models.Todo, error) {
	row := c.DB.QueryRow(`
		UPDATE todos
		SET assigned_to = $1,
		    assigned_by = $2,
		    updated_at  = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, title, description, completed, priority, status,
		          assigned_to, assigned_by, created_at, updated_at
	`, assignedToUserID, assignedByUserID, todoID)

	todo, err := scanTodoRow(row, false)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo with ID %d not found", todoID)
	}
	if err != nil {
		return nil, fmt.Errorf("error assigning todo: %w", err)
	}
	return todo, nil
}

// UpdateStatus advances the workflow stage of a todo.
// Members can update their own assigned todos; admins can update any.
func (c *Connector) UpdateStatus(todoID int, status models.WorkflowStatus) (*models.Todo, error) {
	// completed mirrors status == done
	completed := status == models.StatusDone

	row := c.DB.QueryRow(`
		UPDATE todos
		SET status     = $1,
		    completed  = $2,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, title, description, completed, priority, status,
		          assigned_to, assigned_by, created_at, updated_at
	`, status, completed, todoID)

	todo, err := scanTodoRow(row, false)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo with ID %d not found", todoID)
	}
	if err != nil {
		return nil, fmt.Errorf("error updating status: %w", err)
	}
	return todo, nil
}