package controllers

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/gtkrshnaaa/multicamobserver/internal/models"
)

// BaseController encapsulates shared dependencies for all handlers
type BaseController struct {
	DB             *sql.DB
	JWTSecret      []byte
	Templates      *template.Template
	StreamRegistry *models.StreamRegistry
}

// NewBaseController initializes the base controller with dependencies
func NewBaseController(db *sql.DB, jwtSecret []byte, tmpl *template.Template, registry *models.StreamRegistry) *BaseController {
	return &BaseController{
		DB:             db,
		JWTSecret:      jwtSecret,
		Templates:      tmpl,
		StreamRegistry: registry,
	}
}

// Render is a helper to render Server-Side HTML templates easily
func (c *BaseController) Render(w http.ResponseWriter, r *http.Request, tmplName string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.Templates.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
	}
}

// JSONResponse sends a JSON structured payload back to the client
func (c *BaseController) JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// For simplicity, we just use a naive JSON encoder or standard errors
}
