package services

import (
	"fmt"
	"strings"

	ifaces "github.com/Clivet-lug/todo-app/backend/interfaces"
	"github.com/Clivet-lug/todo-app/backend/models"
)

// todoService holds all todo business logic.
type todoService struct {
	todoRepo ifaces.TodoRepository
	userRepo ifaces.UserRepository
}

func NewTodoService(todoRepo ifaces.TodoRepository, userRepo ifaces.UserRepository) ifaces.TodoService {
	return &todoService{todoRepo: todoRepo, userRepo: userRepo}
}

func (s *todoService) GetTodos(userID int, isAdmin bool) ([]*models.Todo, error) {
	todos, err := s.todoRepo.GetTodos(userID, isAdmin)
	if todos == nil {
		todos = []*models.Todo{}
	}
	return todos, err
}

// GetTodoByID enforces ownership — members can only see their own todos.
func (s *todoService) GetTodoByID(id, requestingUserID int, isAdmin bool) (*models.Todo, error) {
	todo, err := s.todoRepo.GetTodoByID(id)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		if todo.AssignedTo == nil || *todo.AssignedTo != requestingUserID {
			return nil, fmt.Errorf("access denied")
		}
	}
	return todo, nil
}

func (s *todoService) CreateTodo(title, description, priority string) (*models.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if priority == "" {
		priority = "medium"
	}
	return s.todoRepo.CreateTodo(title, strings.TrimSpace(description), priority)
}

func (s *todoService) UpdateTodo(id int, title, description, priority string) (*models.Todo, error) {
	return s.todoRepo.UpdateTodo(id, title, description, priority)
}

func (s *todoService) DeleteTodo(id int) error {
	return s.todoRepo.DeleteTodo(id)
}

// AssignTodo enforces the rule: todos can only be assigned to members, not admins.
func (s *todoService) AssignTodo(todoID, assignedToUserID, assignedByAdminID int) (*models.Todo, error) {
	targetUser, err := s.userRepo.FindByID(assignedToUserID)
	if err != nil {
		return nil, fmt.Errorf("target user not found")
	}
	if targetUser.Role != "member" {
		return nil, fmt.Errorf("todos can only be assigned to members")
	}
	return s.todoRepo.AssignTodo(todoID, assignedToUserID, assignedByAdminID)
}

// UpdateStatus enforces ownership for members and validates the status value.
func (s *todoService) UpdateStatus(todoID, requestingUserID int, isAdmin bool, status models.WorkflowStatus) (*models.Todo, error) {
	if !models.ValidStatuses[status] {
		return nil, fmt.Errorf("status must be one of: todo, in_progress, review, done")
	}
	if !isAdmin {
		todo, err := s.todoRepo.GetTodoByID(todoID)
		if err != nil {
			return nil, err
		}
		if todo.AssignedTo == nil || *todo.AssignedTo != requestingUserID {
			return nil, fmt.Errorf("you can only update status of todos assigned to you")
		}
	}
	return s.todoRepo.UpdateStatus(todoID, status)
}

func (s *todoService) ListMembers() ([]*models.User, error) {
	members, err := s.userRepo.ListMembers()
	if members == nil {
		members = []*models.User{}
	}
	return members, err
}