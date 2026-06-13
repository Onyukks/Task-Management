package task

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tech4mation/tasks-api/internal/events"
	"github.com/tech4mation/tasks-api/internal/httpx"
	"github.com/tech4mation/tasks-api/internal/middleware"
)

const (
	maxPageSize     = 100
	defaultPageSize = 20
)

var (
	validStatuses   = map[string]bool{"todo": true, "in_progress": true, "done": true}
	validPriorities = map[string]bool{"low": true, "medium": true, "high": true}
)

type Handler struct {
	repo   *Repo
	broker *events.Broker
}

func NewHandler(repo *Repo, broker *events.Broker) *Handler {
	return &Handler{repo: repo, broker: broker}
}

// ---- Create ----

type createRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=200"`
	Description string     `json:"description" validate:"max=2000"`
	Status      string     `json:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority    string     `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"dueDate"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())
	var req createRequest
	if !httpx.DecodeAndValidate(w, r, &req) {
		return
	}
	if req.Status == "" {
		req.Status = "todo"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	t, err := h.repo.Create(r.Context(), p.UserID, CreateInput{
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	})
	if err != nil {
		httpx.Internal(w, err)
		return
	}
	// Record the creation in the activity log (best-effort).
	if err := h.repo.logActivities(r.Context(), t.ID, p.UserID, []activityEntry{
		{Action: "created", Detail: "Created the task"},
	}); err != nil {
		slog.Warn("failed to log create activity", "task", t.ID, "error", err)
	}
	h.broker.Publish(p.UserID, events.Event{Type: "task.created", Data: t})
	httpx.JSON(w, http.StatusCreated, t)
}

// ---- List ----

type listResponse struct {
	Tasks      []Task `json:"tasks"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalPages int    `json:"totalPages"`
}

// list handles both the user-scoped and admin (all users) variants.
func (h *Handler) list(w http.ResponseWriter, r *http.Request, allUsers bool, userID uuid.UUID) {
	q := r.URL.Query()

	status := q.Get("status")
	if status != "" && !validStatuses[status] {
		httpx.BadRequest(w, "Invalid status filter.")
		return
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	params := ListParams{
		Status:   status,
		Search:   q.Get("search"),
		SortBy:   q.Get("sortBy"),
		SortDir:  q.Get("sortDir"),
		Page:     page,
		PageSize: pageSize,
		AllUsers: allUsers,
		OnlyUser: userID,
	}

	tasks, total, err := h.repo.List(r.Context(), params)
	if err != nil {
		httpx.Internal(w, err)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	httpx.JSON(w, http.StatusOK, listResponse{
		Tasks:      tasks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())
	h.list(w, r, false, p.UserID)
}

// ListAll is the admin-only endpoint that returns every user's tasks.
func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, true, uuid.Nil)
}

// ---- Get ----

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	t, err := h.repo.Get(r.Context(), id, ownerScope(p))
	if err != nil {
		notFoundOrInternal(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, t)
}

// ---- Update (PATCH) ----

type updateRequest struct {
	Title       *string                   `json:"title"`
	Description *string                   `json:"description"`
	Status      *string                   `json:"status"`
	Priority    *string                   `json:"priority"`
	DueDate     httpx.Optional[time.Time] `json:"dueDate"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req updateRequest
	if !httpx.DecodeAndValidate(w, r, &req) {
		return
	}

	// Manual validation: every field is optional on PATCH, but when present it
	// must be valid.
	fieldErrs := map[string]string{}
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" || len(t) > 200 {
			fieldErrs["title"] = "Must be between 1 and 200 characters."
		} else {
			req.Title = &t
		}
	}
	if req.Description != nil && len(*req.Description) > 2000 {
		fieldErrs["description"] = "Must be at most 2000 characters."
	}
	if req.Status != nil && !validStatuses[*req.Status] {
		fieldErrs["status"] = "Must be one of: todo, in_progress, done."
	}
	if req.Priority != nil && !validPriorities[*req.Priority] {
		fieldErrs["priority"] = "Must be one of: low, medium, high."
	}
	if len(fieldErrs) > 0 {
		httpx.ValidationError(w, fieldErrs)
		return
	}

	in := UpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
	}
	if req.DueDate.Present {
		if req.DueDate.Null {
			in.ClearDue = true
		} else {
			in.DueDate = &req.DueDate.Value
		}
	}

	// Fetch the current state first so we can record exactly what changed.
	before, err := h.repo.Get(r.Context(), id, ownerScope(p))
	if err != nil {
		notFoundOrInternal(w, err)
		return
	}

	t, err := h.repo.Update(r.Context(), id, ownerScope(p), in)
	if err != nil {
		notFoundOrInternal(w, err)
		return
	}

	if entries := diffActivities(before, t); len(entries) > 0 {
		if err := h.repo.logActivities(r.Context(), t.ID, p.UserID, entries); err != nil {
			slog.Warn("failed to log update activity", "task", t.ID, "error", err)
		}
	}
	h.broker.Publish(t.UserID, events.Event{Type: "task.updated", Data: t})
	httpx.JSON(w, http.StatusOK, t)
}

// Activity returns the change history for a task.
func (h *Handler) Activity(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	// Enforce ownership (admins bypass) by loading the task with the same scope.
	if _, err := h.repo.Get(r.Context(), id, ownerScope(p)); err != nil {
		notFoundOrInternal(w, err)
		return
	}
	activity, err := h.repo.ListActivity(r.Context(), id)
	if err != nil {
		httpx.Internal(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"activity": activity})
}

// ---- Delete ----

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.repo.Delete(r.Context(), id, ownerScope(p)); err != nil {
		notFoundOrInternal(w, err)
		return
	}
	h.broker.Publish(p.UserID, events.Event{Type: "task.deleted", Data: map[string]string{"id": id.String()}})
	w.WriteHeader(http.StatusNoContent)
}

// ---- Helpers ----

// ownerScope returns nil for admins (no ownership restriction) and the user's
// id otherwise, so the repo enforces "users only touch their own tasks".
func ownerScope(p middleware.Principal) *uuid.UUID {
	if p.Role == "admin" {
		return nil
	}
	id := p.UserID
	return &id
}

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.BadRequest(w, "Invalid task id.")
		return uuid.Nil, false
	}
	return id, true
}

func notFoundOrInternal(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		httpx.NotFound(w, "Task not found.")
		return
	}
	httpx.Internal(w, err)
}
