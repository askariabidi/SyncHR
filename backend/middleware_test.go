package main

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TC-JWT-01: a token minted by GenerateJWT (the same function Login uses)
// must parse back to the exact claims it was created with.
func TestParseJWT_ValidToken(t *testing.T) {
	tokenString, err := GenerateJWT(42, "jane@example.com", "hr_manager")
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	claims, err := parseJWT(tokenString)
	if err != nil {
		t.Fatalf("parseJWT rejected a token it should have accepted: %v", err)
	}

	if got := int(claims["user_id"].(float64)); got != 42 {
		t.Errorf("user_id = %d, want 42", got)
	}
	if got := claims["email"].(string); got != "jane@example.com" {
		t.Errorf("email = %q, want jane@example.com", got)
	}
	if got := claims["role"].(string); got != "hr_manager" {
		t.Errorf("role = %q, want hr_manager", got)
	}
}

// TC-JWT-02: a token signed with the wrong secret must be rejected -
// this is the core guarantee JWT auth is supposed to provide.
func TestParseJWT_WrongSecret(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id": 1,
		"email":   "attacker@example.com",
		"role":    "hr_manager",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("not_the_real_secret"))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	if _, err := parseJWT(tokenString); err == nil {
		t.Fatal("parseJWT accepted a token signed with the wrong secret")
	}
}

// TC-JWT-03: an expired token must be rejected even if the signature is valid.
func TestParseJWT_ExpiredToken(t *testing.T) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret_key_change_this"
	}

	claims := jwt.MapClaims{
		"user_id": 1,
		"email":   "employee@example.com",
		"role":    "employee",
		"exp":     time.Now().Add(-time.Hour).Unix(), // expired an hour ago
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	if _, err := parseJWT(tokenString); err == nil {
		t.Fatal("parseJWT accepted an expired token")
	}
}

// TC-JWT-04: garbage input must not panic the server, just fail cleanly.
func TestParseJWT_MalformedToken(t *testing.T) {
	if _, err := parseJWT("this.is.not-a-jwt"); err == nil {
		t.Fatal("parseJWT accepted a malformed token")
	}
}

// TC-JWT-05: a token signed with the "none" algorithm (a classic JWT
// vulnerability) must be rejected - parseJWT explicitly checks the signing
// method is HMAC before trusting anything else about the token.
func TestParseJWT_RejectsNoneAlgorithm(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id": 1,
		"email":   "attacker@example.com",
		"role":    "hr_manager",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to sign none-alg test token: %v", err)
	}

	if _, err := parseJWT(tokenString); err == nil {
		t.Fatal("parseJWT accepted a token using the 'none' signing algorithm")
	}
}
