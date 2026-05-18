package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserContextKey contextKey = "user_claims"
)

// AuthClaims represents custom JWT claims
type AuthClaims struct {
	Subject string `json:"sub"` // Email for admin, node_id for broadcaster
	Role    string `json:"role"` // "admin" or "broadcaster"
	jwt.RegisteredClaims
}

// GenerateJWT creates a new stateless JWT token signed with a key
func GenerateJWT(subject, role string, secret []byte) (string, error) {
	claims := AuthClaims{
		Subject: subject,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// RequireAuth is a middleware that validates JWT and requires a specific role
func RequireAuth(roleRequired string, jwtSecret []byte, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			// No token provided, redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tokenString := cookie.Value
		claims := &AuthClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			// Invalid token, clear cookie and redirect
			clearCookie(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Verify role matches
		if claims.Role != roleRequired {
			http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
			return
		}

		// Store claims in context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// GetUserClaims retrieves parsed JWT claims from request context
func GetUserClaims(r *http.Request) *AuthClaims {
	claims, ok := r.Context().Value(UserContextKey).(*AuthClaims)
	if !ok {
		return nil
	}
	return claims
}

func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // Set to true in prod/HTTPS environment
		SameSite: http.SameSiteLaxMode,
	})
}
