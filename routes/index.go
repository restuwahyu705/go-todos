package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sqlb "github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
)

type Todos struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Category    string    `json:"category" db:"category"`
	Description *string   `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type APIResponse struct {
	StatCode    int         `json:"stat_code"`
	StatMessage string      `json:"stat_message"`
	Data        interface{} `json:"data"`
}

type routesInterface interface {
	TodosRouter(w http.ResponseWriter, r *http.Request)
}

type routesStruct struct {
	db *sqlx.DB
}

func NewRouter(db *sqlx.DB) routesInterface {
	return &routesStruct{db: db}
}

func (h *routesStruct) TodosRouter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	todos := []Todos{}
	todo := Todos{}
	psql := sqlb.PostgreSQL

	if r.Method == http.MethodGet {
		query := psql.NewSelectBuilder().Select("*").From("todos").String()

		ctx, close := context.WithTimeout(r.Context(), time.Duration(time.Second*5))
		defer close()

		if err := h.db.SelectContext(ctx, &todos, query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&APIResponse{
				StatCode:    http.StatusInternalServerError,
				StatMessage: fmt.Sprintf("Sql Query Error: %s", err.Error()),
				Data:        todos,
			})
			return
		}

		if len(todos) < 1 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&APIResponse{
				StatCode:    http.StatusNotFound,
				StatMessage: "Todos data already not found",
				Data:        todos,
			})
			return
		}

		json.NewEncoder(w).Encode(&APIResponse{
			StatCode:    http.StatusOK,
			StatMessage: "Todos data already to use",
			Data:        todos,
		})
	} else if r.Method == http.MethodPost {
		query := psql.NewInsertBuilder().InsertInto("todos").Cols("name", "category", "description").Values("?", "?", "?").String()

		if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&APIResponse{
				StatCode:    http.StatusBadRequest,
				StatMessage: fmt.Sprintf("Request body not valid: %s", err.Error()),
			})
			return
		}

		ctx, close := context.WithTimeout(r.Context(), time.Duration(time.Second*3))
		defer close()

		if _, err := h.db.ExecContext(ctx, query, todo.Name, todo.Category, todo.Description); err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(&APIResponse{
				StatCode:    http.StatusForbidden,
				StatMessage: fmt.Sprintf("Insert new todos failed: %s", err.Error()),
			})
			return
		}

		json.NewEncoder(w).Encode(&APIResponse{
			StatCode:    http.StatusOK,
			StatMessage: "Insert new todos success",
		})
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(&APIResponse{
			StatCode:    http.StatusMethodNotAllowed,
			StatMessage: "Request method not allowed",
		})
	}
}
