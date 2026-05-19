package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	ifaces "github.com/Clivet-lug/todo-app/backend/interfaces"
	"github.com/Clivet-lug/todo-app/backend/middleware"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/utils"
)

type TodoHandler struct {
	svc ifaces.TodoService
}

func NewTodoHandler(svc ifaces.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	todos, err := h.svc.GetTodos(claims.UserID, claims.Role == "admin")
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Error fetching todos"})
		return
	}
	utils.SendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Todos retrieved", Data: todos, Count: len(todos)})
}

func (h *TodoHandler) GetTodoByID(w http.ResponseWriter, r *http.Request, id int) {
	claims := middleware.GetClaims(r)
	todo, err := h.svc.GetTodoByID(id, claims.UserID, claims.Role == "admin")
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "access denied" {
			status = http.StatusForbidden
		} else if err.Error() == fmt.Sprintf("todo %d not found", id) {
			status = http.StatusNotFound
		}
		utils.SendJSON(w, status, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	utils.SendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Todo retrieved", Data: todo})
}

func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var input models.Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request body"})
		return
	}
	todo, err := h.svc.CreateTodo(input.Title, input.Description, input.Priority)
	if err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	utils.SendJSON(w, http.StatusCreated, models.APIResponse{Success: true, Message: "Todo created", Data: todo})
}

func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request, id int) {
	var input models.Todo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request body"})
		return
	}
	todo, err := h.svc.UpdateTodo(id, input.Title, input.Description, input.Priority)
	if err != nil {
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	utils.SendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Todo updated", Data: todo})
}

func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	if err := h.svc.DeleteTodo(id); err != nil {
		utils.SendJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	utils.SendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Todo deleted"})
}

func (h *TodoHandler) AssignTodo(w http.ResponseWriter, r *http.Request, id int) {
	claims := middleware.GetClaims(r)
	var body struct {
		AssignedTo int `json:"assigned_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AssignedTo == 0 {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "assigned_to (user ID) is required"})
		return
	}
	todo, err := h.svc.AssignTodo(id, body.AssignedTo, claims.UserID)
	if err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	utils.SendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Todo assigned", Data: todo})
}

func (h *TodoHandler) UpdateStatus(w http.ResponseWriter, r *http.Request, id int) {
	claims := middleware.GetClaims(r)
	var body struct {
		Status models.WorkflowStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request body"})
		return
	}
	todo, err := h.svc.UpdateStatus(id, claims.UserID, claims.Role == "admin", body.Status)
	if err != nil {
		utils.SendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	utils.SendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Status updated", Data: todo})
}

func (h *TodoHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.svc.ListMembers()
	if err != nil {
		utils.SendJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Error fetching members"})
		return
	}
	utils.SendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Members retrieved", Data: members, Count: len(members)})
}

// GetTodosData is used by the server's cache layer to get raw data
// without writing an HTTP response directly.
func (h *TodoHandler) GetTodosData(userID int, isAdmin bool) (interface{}, error) {
	return h.svc.GetTodos(userID, isAdmin)
}