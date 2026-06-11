package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if hash == "correct-horse-battery" {
		t.Fatal("password was stored in plaintext")
	}
	if !CheckPassword(hash, "correct-horse-battery") {
		t.Error("correct password rejected")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("wrong password accepted")
	}
}

func TestJWTRoundTrip(t *testing.T) {
	issuer := NewIssuer([]byte("a-test-secret-that-is-long-enough-32"), time.Hour)
	uid := uuid.New()

	token, expiresAt, err := issuer.Issue(uid, "admin")
	if err != nil {
		t.Fatalf("issue failed: %v", err)
	}
	if !expiresAt.After(time.Now()) {
		t.Error("expiry should be in the future")
	}

	claims, err := issuer.Verify(token)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if claims.UserID != uid {
		t.Errorf("got user id %s, want %s", claims.UserID, uid)
	}
	if claims.Role != "admin" {
		t.Errorf("got role %q, want admin", claims.Role)
	}
}

func TestJWTRejectsTamperedAndForeignTokens(t *testing.T) {
	issuer := NewIssuer([]byte("a-test-secret-that-is-long-enough-32"), time.Hour)
	token, _, _ := issuer.Issue(uuid.New(), "user")

	// Tamper with the payload.
	if _, err := issuer.Verify(token + "x"); err == nil {
		t.Error("tampered token was accepted")
	}

	// Token signed with a different secret must be rejected.
	other := NewIssuer([]byte("a-completely-different-secret-32chars"), time.Hour)
	foreign, _, _ := other.Issue(uuid.New(), "user")
	if _, err := issuer.Verify(foreign); err == nil {
		t.Error("token signed by a different secret was accepted")
	}

	// Expired token must be rejected.
	expiredIssuer := NewIssuer([]byte("a-test-secret-that-is-long-enough-32"), -time.Hour)
	expired, _, _ := expiredIssuer.Issue(uuid.New(), "user")
	if _, err := issuer.Verify(expired); err == nil {
		t.Error("expired token was accepted")
	}
}
