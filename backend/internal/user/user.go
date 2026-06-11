// Package user owns user persistence and the account model.
package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// normalizeEmail lower-cases and trims an email so lookups are case-insensitive.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

var ErrEmailTaken = errors.New("email already registered")
var ErrNotFound = errors.New("user not found")

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	PasswordHash string    `json:"-"` // never serialized
	CreatedAt    time.Time `json:"createdAt"`
}

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

// Create inserts a new user. Returns ErrEmailTaken on a unique-violation.
func (r *Repo) Create(ctx context.Context, email, name, passwordHash string) (*User, error) {
	const q = `
		INSERT INTO users (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, role, created_at`
	var u User
	err := r.db.QueryRow(ctx, q, email, name, passwordHash).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return &u, nil
}

// ByEmail loads a user (including the password hash) for login.
func (r *Repo) ByEmail(ctx context.Context, email string) (*User, error) {
	const q = `SELECT id, email, name, role, password_hash, created_at FROM users WHERE email = $1`
	var u User
	err := r.db.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

// ByID loads a user by id (used to hydrate the session on the frontend).
func (r *Repo) ByID(ctx context.Context, id uuid.UUID) (*User, error) {
	const q = `SELECT id, email, name, role, password_hash, created_at FROM users WHERE id = $1`
	var u User
	err := r.db.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}
