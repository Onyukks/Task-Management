package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/tech4mation/tasks-api/internal/auth"
	"github.com/tech4mation/tasks-api/internal/config"
	"github.com/tech4mation/tasks-api/internal/httpx"
	"github.com/tech4mation/tasks-api/internal/middleware"
)

// Handler serves the auth endpoints.
type Handler struct {
	repo   *Repo
	issuer *auth.Issuer
	cfg    *config.Config
}

func NewHandler(repo *Repo, issuer *auth.Issuer, cfg *config.Config) *Handler {
	return &Handler{repo: repo, issuer: issuer, cfg: cfg}
}

type signupRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Name     string `json:"name" validate:"required,min=2,max=80"`
	Password string `json:"password" validate:"required,min=8,max=72"` // bcrypt caps at 72 bytes
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type authResponse struct {
	User *User `json:"user"`
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if !httpx.DecodeAndValidate(w, r, &req) {
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		httpx.Internal(w, err)
		return
	}

	u, err := h.repo.Create(r.Context(), normalizeEmail(req.Email), req.Name, hash)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			httpx.Conflict(w, "An account with that email already exists.")
			return
		}
		httpx.Internal(w, err)
		return
	}

	h.setSessionCookie(w, u)
	httpx.JSON(w, http.StatusCreated, authResponse{User: u})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !httpx.DecodeAndValidate(w, r, &req) {
		return
	}

	u, err := h.repo.ByEmail(r.Context(), normalizeEmail(req.Email))
	// Run the password check even when the user is missing to keep timing
	// roughly constant and avoid leaking which emails exist.
	if err != nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		httpx.Unauthorized(w, "Invalid email or password.")
		return
	}

	h.setSessionCookie(w, u)
	httpx.JSON(w, http.StatusOK, authResponse{User: u})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearSessionCookie(w)
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me returns the currently authenticated user (used to rehydrate the session
// after a page refresh on the frontend).
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())
	u, err := h.repo.ByID(r.Context(), p.UserID)
	if err != nil {
		httpx.Unauthorized(w, "Session no longer valid.")
		return
	}
	httpx.JSON(w, http.StatusOK, authResponse{User: u})
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, u *User) {
	token, expiresAt, err := h.issuer.Issue(u.ID, u.Role)
	if err != nil {
		httpx.Internal(w, err)
		return
	}
	sameSite := http.SameSiteLaxMode
	if h.cfg.Production {
		// Cross-site (frontend and API on different domains) requires None+Secure.
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.cfg.Production,
		SameSite: sameSite,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	sameSite := http.SameSiteLaxMode
	if h.cfg.Production {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.Production,
		SameSite: sameSite,
	})
}
