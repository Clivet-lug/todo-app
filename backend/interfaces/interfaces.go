package interfaces

import "github.com/Clivet-lug/todo-app/backend/models"

// REPOSITORY INTERFACES
// Repositories only speak SQL.

type UserRepository interface {
	CreateUsersTable() error
	CreateUser(name, email, hashedPassword, role string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error)
	ListMembers() ([]*models.User, error)
}

type TodoRepository interface {
	CreateTable() error
	GetTodos(userID int, isAdmin bool) ([]*models.Todo, error)
	GetTodoByID(id int) (*models.Todo, error)
	CreateTodo(title, description, priority string) (*models.Todo, error)
	UpdateTodo(id int, title, description, priority string) (*models.Todo, error)
	DeleteTodo(id int) error
	AssignTodo(todoID, assignedToUserID, assignedByUserID int) (*models.Todo, error)
	UpdateStatus(todoID int, status models.WorkflowStatus) (*models.Todo, error)
}

// SERVICE INTERFACES
// Services hold business logic. Handlers depend on these interfaces.
type AuthService interface {
	Register(name, email, password, role string) (*models.User, string, error)
	Login(email, password string) (*models.User, string, error)
}

type TodoService interface {
	GetTodos(userID int, isAdmin bool) ([]*models.Todo, error)
	GetTodoByID(id, requestingUserID int, isAdmin bool) (*models.Todo, error)
	CreateTodo(title, description, priority string) (*models.Todo, error)
	UpdateTodo(id int, title, description, priority string) (*models.Todo, error)
	DeleteTodo(id int) error
	AssignTodo(todoID, assignedToUserID, assignedByAdminID int) (*models.Todo, error)
	UpdateStatus(todoID, requestingUserID int, isAdmin bool, status models.WorkflowStatus) (*models.Todo, error)
	ListMembers() ([]*models.User, error)
}