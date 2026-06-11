// Package task owns task persistence, the task model, and list querying.
package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"userId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"dueDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// ListParams captures every filter/search/sort/pagination input. They are
// designed to compose: a status filter, a title search, and a sort can all be
// applied at once.
type ListParams struct {
	Status   string // "" = any
	Search   string // title contains (case-insensitive)
	SortBy   string // due_date | priority | created_at
	SortDir  string // asc | desc
	Page     int    // 1-based
	PageSize int

	// When set, list across all users (admin). Otherwise scoped to OnlyUser.
	AllUsers bool
	OnlyUser uuid.UUID
}

// sortColumns whitelists the columns a client may sort by, mapping the API
// name to a safe SQL expression. This prevents SQL injection via sort input.
var sortColumns = map[string]string{
	"due_date":   "due_date",
	"priority":   "priority", // enum ordered low < medium < high by declaration
	"created_at": "created_at",
}

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

type CreateInput struct {
	Title       string
	Description string
	Status      string
	Priority    string
	DueDate     *time.Time
}

func (r *Repo) Create(ctx context.Context, userID uuid.UUID, in CreateInput) (*Task, error) {
	const q = `
		INSERT INTO tasks (user_id, title, description, status, priority, due_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, title, description, status, priority, due_date, created_at, updated_at`
	var t Task
	err := r.db.QueryRow(ctx, q, userID, in.Title, in.Description, in.Status, in.Priority, in.DueDate).
		Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	return &t, err
}

// Get fetches a single task. If onlyUser is non-nil, the task must belong to
// that user or ErrNotFound is returned (avoids leaking existence to others).
func (r *Repo) Get(ctx context.Context, id uuid.UUID, onlyUser *uuid.UUID) (*Task, error) {
	q := `SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at
	      FROM tasks WHERE id = $1`
	args := []any{id}
	if onlyUser != nil {
		q += ` AND user_id = $2`
		args = append(args, *onlyUser)
	}
	var t Task
	err := r.db.QueryRow(ctx, q, args...).
		Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &t, err
}

// UpdateInput holds optional fields; nil pointers are left unchanged (PATCH).
type UpdateInput struct {
	Title       *string
	Description *string
	Status      *string
	Priority    *string
	DueDate     *time.Time
	ClearDue    bool // explicitly set due_date to NULL
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, ownerScope *uuid.UUID, in UpdateInput) (*Task, error) {
	set := []string{}
	args := []any{}
	i := 1
	add := func(expr string, val any) {
		set = append(set, fmt.Sprintf("%s = $%d", expr, i))
		args = append(args, val)
		i++
	}

	if in.Title != nil {
		add("title", *in.Title)
	}
	if in.Description != nil {
		add("description", *in.Description)
	}
	if in.Status != nil {
		add("status", *in.Status)
	}
	if in.Priority != nil {
		add("priority", *in.Priority)
	}
	if in.ClearDue {
		set = append(set, "due_date = NULL")
	} else if in.DueDate != nil {
		add("due_date", *in.DueDate)
	}

	if len(set) == 0 {
		// Nothing to change; just return the current row.
		return r.Get(ctx, id, ownerScope)
	}
	set = append(set, "updated_at = now()")

	q := "UPDATE tasks SET " + strings.Join(set, ", ") + fmt.Sprintf(" WHERE id = $%d", i)
	args = append(args, id)
	i++
	if ownerScope != nil {
		q += fmt.Sprintf(" AND user_id = $%d", i)
		args = append(args, *ownerScope)
	}
	q += " RETURNING id, user_id, title, description, status, priority, due_date, created_at, updated_at"

	var t Task
	err := r.db.QueryRow(ctx, q, args...).
		Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &t, err
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID, ownerScope *uuid.UUID) error {
	q := "DELETE FROM tasks WHERE id = $1"
	args := []any{id}
	if ownerScope != nil {
		q += " AND user_id = $2"
		args = append(args, *ownerScope)
	}
	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// List returns a page of tasks plus the total count matching the filters.
func (r *Repo) List(ctx context.Context, p ListParams) ([]Task, int, error) {
	where := []string{}
	args := []any{}
	i := 1
	addArg := func(v any) string {
		args = append(args, v)
		ph := fmt.Sprintf("$%d", i)
		i++
		return ph
	}

	if !p.AllUsers {
		where = append(where, "user_id = "+addArg(p.OnlyUser))
	}
	if p.Status != "" {
		where = append(where, "status = "+addArg(p.Status))
	}
	if s := strings.TrimSpace(p.Search); s != "" {
		where = append(where, "title ILIKE "+addArg("%"+s+"%"))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	// Total count for pagination metadata.
	var total int
	if err := r.db.QueryRow(ctx, "SELECT count(*) FROM tasks"+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	col, ok := sortColumns[p.SortBy]
	if !ok {
		col = "created_at"
	}
	dir := "DESC"
	if strings.ToLower(p.SortDir) == "asc" {
		dir = "ASC"
	}
	// NULLS LAST keeps tasks without a due date at the end regardless of dir.
	orderSQL := fmt.Sprintf(" ORDER BY %s %s NULLS LAST, id %s", col, dir, dir)

	limit := addArg(p.PageSize)
	offset := addArg((p.Page - 1) * p.PageSize)

	q := `SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at
	      FROM tasks` + whereSQL + orderSQL + " LIMIT " + limit + " OFFSET " + offset

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}
