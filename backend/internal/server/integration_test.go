package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tech4mation/tasks-api/internal/config"
	"github.com/tech4mation/tasks-api/internal/db"
	"github.com/tech4mation/tasks-api/internal/server"
)

// These tests run against a real Postgres instance. Set TEST_DATABASE_URL to
// enable them (CI provides one); otherwise they are skipped so `go test ./...`
// stays green on a machine without a database.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	pool, err := db.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := db.Migrate(pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Clean slate for a deterministic test.
	if _, err := pool.Exec(context.Background(), "TRUNCATE tasks, users CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

func newTestServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
	t.Helper()
	cfg := &config.Config{
		JWTSecret:      []byte("integration-test-secret-32-chars-min"),
		JWTTTL:         time.Hour,
		FrontendOrigin: "http://localhost:3000",
	}
	srv := httptest.NewServer(server.New(pool, cfg))
	t.Cleanup(srv.Close)
	return srv
}

// client is a cookie-aware HTTP client representing one logged-in user.
type client struct {
	t   *testing.T
	srv *httptest.Server
	c   *http.Client
}

func newClient(t *testing.T, srv *httptest.Server) *client {
	jar, _ := cookiejar.New(nil)
	return &client{t: t, srv: srv, c: &http.Client{Jar: jar}}
}

func (c *client) do(method, path string, body any) *http.Response {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, err := http.NewRequest(method, c.srv.URL+path, &buf)
	if err != nil {
		c.t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.c.Do(req)
	if err != nil {
		c.t.Fatalf("request %s %s: %v", method, path, err)
	}
	return resp
}

func (c *client) signup(email, name, password string) {
	resp := c.do(http.MethodPost, "/auth/signup", map[string]string{
		"email": email, "name": name, "password": password,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		c.t.Fatalf("signup %s: status %d", email, resp.StatusCode)
	}
}

func (c *client) createTask(title string) string {
	resp := c.do(http.MethodPost, "/tasks/", map[string]any{"title": title})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		c.t.Fatalf("create task: status %d", resp.StatusCode)
	}
	var out struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	return out.ID
}

// TestTaskOwnershipIsolation is the security-critical test: user B must not be
// able to read, update, or delete user A's task, and listing is scoped per user.
func TestTaskOwnershipIsolation(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	srv := newTestServer(t, pool)

	alice := newClient(t, srv)
	alice.signup("alice@example.com", "Alice", "password123")
	taskID := alice.createTask("Alice private task")

	bob := newClient(t, srv)
	bob.signup("bob@example.com", "Bob", "password123")

	// Bob cannot read Alice's task — must look like it does not exist.
	if resp := bob.do(http.MethodGet, "/tasks/"+taskID, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("Bob GET Alice task: got %d, want 404", resp.StatusCode)
	}
	// Bob cannot update it.
	if resp := bob.do(http.MethodPatch, "/tasks/"+taskID, map[string]string{"title": "hijacked"}); resp.StatusCode != http.StatusNotFound {
		t.Errorf("Bob PATCH Alice task: got %d, want 404", resp.StatusCode)
	}
	// Bob cannot delete it.
	if resp := bob.do(http.MethodDelete, "/tasks/"+taskID, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("Bob DELETE Alice task: got %d, want 404", resp.StatusCode)
	}
	// Bob's own list is empty; Alice's has one task.
	if got := bob.listCount("/tasks/"); got != 0 {
		t.Errorf("Bob list count: got %d, want 0", got)
	}
	if got := alice.listCount("/tasks/"); got != 1 {
		t.Errorf("Alice list count: got %d, want 1", got)
	}
}

// TestFilterSearchSortCompose verifies status filter + title search + sort all
// work together on the list endpoint.
func TestFilterSearchSortCompose(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	srv := newTestServer(t, pool)

	u := newClient(t, srv)
	u.signup("composer@example.com", "Composer", "password123")
	u.createTask("Write report")
	u.createTask("Write tests")
	u.createTask("Read book")

	// Search "write" should match exactly two tasks.
	if got := u.listCount("/tasks/?search=write"); got != 2 {
		t.Errorf("search=write count: got %d, want 2", got)
	}
	// Search + status filter compose: none of the "write" tasks are done.
	if got := u.listCount("/tasks/?search=write&status=done"); got != 0 {
		t.Errorf("search=write&status=done count: got %d, want 0", got)
	}
	// Sort by created_at asc returns the first-created task first.
	resp := u.do(http.MethodGet, "/tasks/?sortBy=created_at&sortDir=asc", nil)
	defer resp.Body.Close()
	var list struct {
		Tasks []struct {
			Title string `json:"title"`
		} `json:"tasks"`
	}
	json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Tasks) == 0 || list.Tasks[0].Title != "Write report" {
		t.Errorf("sort asc first task: got %+v, want 'Write report' first", list.Tasks)
	}
}

func (c *client) listCount(path string) int {
	resp := c.do(http.MethodGet, path, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.t.Fatalf("list %s: status %d", path, resp.StatusCode)
	}
	var out struct {
		Total int `json:"total"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	return out.Total
}
