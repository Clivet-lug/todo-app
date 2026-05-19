package repositories

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Clivet-lug/todo-app/backend/models"
)

type todoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) *todoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) CreateTable() error {
	_, err := r.db.Exec(`
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
	`)
	if err != nil {
		log.Print("Error creating todos table:", err)
		return err
	}
	fmt.Println("✅ Todos table ready")
	return r.migrate()
}

func (r *todoRepository) migrate() error {
	migrations := []string{
		`DO $$ BEGIN ALTER TABLE todos ADD COLUMN status VARCHAR(50) DEFAULT 'todo';
		 EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE todos ADD COLUMN assigned_to INT REFERENCES users(id) ON DELETE SET NULL;
		 EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE todos ADD COLUMN assigned_by INT REFERENCES users(id) ON DELETE SET NULL;
		 EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
	}
	for _, m := range migrations {
		if _, err := r.db.Exec(m); err != nil {
			return err
		}
	}
	fmt.Println("✅ Todos migrations applied")
	return nil
}

// scanTodo handles rows that include an optional LEFT JOIN on users.
func scanTodo(row interface{ Scan(...interface{}) error }, withAssignee bool) (*models.Todo, error) {
	var todo models.Todo
	var assignedTo, assignedBy sql.NullInt64

	if withAssignee {
		var aName, aEmail sql.NullString
		err := row.Scan(
			&todo.ID, &todo.Title, &todo.Description, &todo.Completed,
			&todo.Priority, &todo.Status, &assignedTo, &assignedBy,
			&todo.CreatedAt, &todo.UpdatedAt, &aName, &aEmail,
		)
		if err != nil {
			return nil, err
		}
		if assignedTo.Valid {
			v := int(assignedTo.Int64); todo.AssignedTo = &v
		}
		if assignedBy.Valid {
			v := int(assignedBy.Int64); todo.AssignedBy = &v
		}
		if aName.Valid {
			todo.Assignee = &models.UserSummary{ID: int(assignedTo.Int64), Name: aName.String, Email: aEmail.String}
		}
		return &todo, nil
	}

	err := row.Scan(
		&todo.ID, &todo.Title, &todo.Description, &todo.Completed,
		&todo.Priority, &todo.Status, &assignedTo, &assignedBy,
		&todo.CreatedAt, &todo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if assignedTo.Valid { v := int(assignedTo.Int64); todo.AssignedTo = &v }
	if assignedBy.Valid { v := int(assignedBy.Int64); todo.AssignedBy = &v }
	return &todo, nil
}

func (r *todoRepository) GetTodos(userID int, isAdmin bool) ([]*models.Todo, error) {
	base := `
		SELECT t.id, t.title, t.description, t.completed, t.priority,
		       t.status, t.assigned_to, t.assigned_by, t.created_at, t.updated_at,
		       u.name, u.email
		FROM todos t LEFT JOIN users u ON u.id = t.assigned_to`

	var rows *sql.Rows
	var err error
	if isAdmin {
		rows, err = r.db.Query(base + ` ORDER BY t.created_at DESC`)
	} else {
		rows, err = r.db.Query(base+` WHERE t.assigned_to = $1 ORDER BY t.created_at DESC`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*models.Todo
	for rows.Next() {
		t, err := scanTodo(rows, true)
		if err != nil {
			continue
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (r *todoRepository) GetTodoByID(id int) (*models.Todo, error) {
	row := r.db.QueryRow(`
		SELECT t.id, t.title, t.description, t.completed, t.priority,
		       t.status, t.assigned_to, t.assigned_by, t.created_at, t.updated_at,
		       u.name, u.email
		FROM todos t LEFT JOIN users u ON u.id = t.assigned_to
		WHERE t.id = $1`, id)
	t, err := scanTodo(row, true)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo %d not found", id)
	}
	return t, err
}

func (r *todoRepository) CreateTodo(title, description, priority string) (*models.Todo, error) {
	row := r.db.QueryRow(`
		INSERT INTO todos (title, description, priority, status)
		VALUES ($1, $2, $3, 'todo')
		RETURNING id, title, description, completed, priority, status,
		          assigned_to, assigned_by, created_at, updated_at`,
		title, description, priority)
	return scanTodo(row, false)
}

func (r *todoRepository) UpdateTodo(id int, title, description, priority string) (*models.Todo, error) {
	row := r.db.QueryRow(`
		UPDATE todos
		SET title       = COALESCE(NULLIF($1,''), title),
		    description = COALESCE(NULLIF($2,''), description),
		    priority    = COALESCE(NULLIF($3,''), priority),
		    updated_at  = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, title, description, completed, priority, status,
		          assigned_to, assigned_by, created_at, updated_at`,
		title, description, priority, id)
	t, err := scanTodo(row, false)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo %d not found", id)
	}
	return t, err
}

func (r *todoRepository) DeleteTodo(id int) error {
	res, err := r.db.Exec(`DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("todo %d not found", id)
	}
	return nil
}

func (r *todoRepository) AssignTodo(todoID, assignedToUserID, assignedByUserID int) (*models.Todo, error) {
	row := r.db.QueryRow(`
		UPDATE todos SET assigned_to=$1, assigned_by=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id=$3
		RETURNING id, title, description, completed, priority, status,
		          assigned_to, assigned_by, created_at, updated_at`,
		assignedToUserID, assignedByUserID, todoID)
	t, err := scanTodo(row, false)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo %d not found", todoID)
	}
	return t, err
}

func (r *todoRepository) UpdateStatus(todoID int, status models.WorkflowStatus) (*models.Todo, error) {
	row := r.db.QueryRow(`
		UPDATE todos SET status=$1, completed=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id=$3
		RETURNING id, title, description, completed, priority, status,
		          assigned_to, assigned_by, created_at, updated_at`,
		status, status == models.StatusDone, todoID)
	t, err := scanTodo(row, false)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo %d not found", todoID)
	}
	return t, err
}