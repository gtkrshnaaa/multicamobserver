package controllers

import (
	"net/http"
	"time"

	"github.com/gtkrshnaaa/multicamobserver/internal/middleware"
	"github.com/gtkrshnaaa/multicamobserver/internal/models"
)

// ShowLogin renders the unified login page
func (c *BaseController) ShowLogin(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect based on their role
	if cookie, err := r.Cookie("auth_token"); err == nil {
		_ = cookie // check token if wanted, otherwise just render login
	}
	c.Render(w, r, "login.html", nil)
}

// HandleLogin processes credentials submitted via POST
func (c *BaseController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		c.Render(w, r, "login.html", map[string]string{"Error": "Invalid form submission"})
		return
	}

	role := r.FormValue("role")
	username := r.FormValue("username") // Email for admin, NodeID for broadcaster
	password := r.FormValue("password")

	var subject string
	var redirectURL string

	if role == "admin" {
		user, err := models.AuthenticateUser(c.DB, username, password)
		if err != nil {
			c.Render(w, r, "login.html", map[string]string{"Error": "Invalid email or password"})
			return
		}
		subject = user.Email
		redirectURL = "/admin/dashboard"
	} else if role == "broadcaster" {
		broadcaster, err := models.AuthenticateBroadcaster(c.DB, username, password)
		if err != nil {
			c.Render(w, r, "login.html", map[string]string{"Error": "Invalid camera node credentials"})
			return
		}
		subject = broadcaster.NodeID
		redirectURL = "/broadcaster/camera?node_id=" + broadcaster.NodeID
	} else {
		c.Render(w, r, "login.html", map[string]string{"Error": "Invalid role selected"})
		return
	}

	// Generate signed JWT token
	tokenString, err := middleware.GenerateJWT(subject, role, c.JWTSecret)
	if err != nil {
		c.Render(w, r, "login.html", map[string]string{"Error": "Failed to generate session"})
		return
	}

	// Store JWT in secure HttpOnly Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Set to true if running over HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// HandleLogout clears the authorization cookie and redirects to login
func (c *BaseController) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
