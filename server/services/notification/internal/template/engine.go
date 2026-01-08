// Package template provides a dynamic template engine for notifications.
// Templates are stored in the database and compiled on-demand with caching.
package template

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sync"
	"text/template"
	"time"
)

// Template represents a notification template stored in the database.
type Template struct {
	ID        string
	Name      string // Unique name for lookup (e.g., "booking_confirmation")
	Subject   string // Email subject or SMS title (templated)
	Body      string // Template body (Go template syntax)
	Channel   string // "email", "sms", "push"
	Locale    string // "en", "bn", etc.
	Version   int    // For versioning
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Engine is the template engine that manages template loading and rendering.
type Engine struct {
	db       *sql.DB
	cache    map[string]*compiledTemplate
	mu       sync.RWMutex
	cacheTTL time.Duration
}

type compiledTemplate struct {
	tmpl        *template.Template
	subjectTmpl *template.Template
	cachedAt    time.Time
}

// NewEngine creates a new template engine.
func NewEngine(db *sql.DB) *Engine {
	return &Engine{
		db:       db,
		cache:    make(map[string]*compiledTemplate),
		cacheTTL: 5 * time.Minute,
	}
}

// cacheKey generates a cache key for a template.
func cacheKey(name, channel, locale string) string {
	return fmt.Sprintf("%s:%s:%s", name, channel, locale)
}

// GetTemplate retrieves a template from the database.
func (e *Engine) GetTemplate(ctx context.Context, name, channel, locale string) (*Template, error) {
	query := `
		SELECT id, name, subject, body, channel, locale, version, is_active, created_at, updated_at
		FROM notification_templates
		WHERE name = $1 AND channel = $2 AND locale = $3 AND is_active = true
		ORDER BY version DESC
		LIMIT 1
	`

	var t Template
	err := e.db.QueryRowContext(ctx, query, name, channel, locale).Scan(
		&t.ID, &t.Name, &t.Subject, &t.Body, &t.Channel, &t.Locale,
		&t.Version, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Fallback to default locale
		if locale != "en" {
			return e.GetTemplate(ctx, name, channel, "en")
		}
		return nil, fmt.Errorf("template not found: %s/%s/%s", name, channel, locale)
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// Render renders a template with the given data.
func (e *Engine) Render(ctx context.Context, name, channel, locale string, data map[string]interface{}) (subject, body string, err error) {
	key := cacheKey(name, channel, locale)

	// Check cache
	e.mu.RLock()
	cached, ok := e.cache[key]
	e.mu.RUnlock()

	if ok && time.Since(cached.cachedAt) < e.cacheTTL {
		return e.executeTemplate(cached, data)
	}

	// Cache miss - load and compile template
	tmpl, err := e.GetTemplate(ctx, name, channel, locale)
	if err != nil {
		return "", "", err
	}

	compiled, err := e.compileTemplate(tmpl)
	if err != nil {
		return "", "", err
	}

	// Cache the compiled template
	e.mu.Lock()
	e.cache[key] = compiled
	e.mu.Unlock()

	return e.executeTemplate(compiled, data)
}

// compileTemplate compiles a template into executable form.
func (e *Engine) compileTemplate(t *Template) (*compiledTemplate, error) {
	bodyTmpl, err := template.New("body").Parse(t.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body template: %w", err)
	}

	subjectTmpl, err := template.New("subject").Parse(t.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subject template: %w", err)
	}

	return &compiledTemplate{
		tmpl:        bodyTmpl,
		subjectTmpl: subjectTmpl,
		cachedAt:    time.Now(),
	}, nil
}

// executeTemplate executes a compiled template with data.
func (e *Engine) executeTemplate(c *compiledTemplate, data map[string]interface{}) (string, string, error) {
	var subjectBuf, bodyBuf bytes.Buffer

	if err := c.subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute subject template: %w", err)
	}

	if err := c.tmpl.Execute(&bodyBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %w", err)
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}

// CreateTemplate creates a new template in the database.
func (e *Engine) CreateTemplate(ctx context.Context, t *Template) error {
	query := `
		INSERT INTO notification_templates (id, name, subject, body, channel, locale, version, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	t.Version = 1
	t.IsActive = true

	_, err := e.db.ExecContext(ctx, query,
		t.ID, t.Name, t.Subject, t.Body, t.Channel, t.Locale,
		t.Version, t.IsActive, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Invalidate cache
	e.mu.Lock()
	delete(e.cache, cacheKey(t.Name, t.Channel, t.Locale))
	e.mu.Unlock()

	return nil
}

// UpdateTemplate updates an existing template (creates a new version).
func (e *Engine) UpdateTemplate(ctx context.Context, t *Template) error {
	// Deactivate old versions
	_, err := e.db.ExecContext(ctx,
		`UPDATE notification_templates SET is_active = false WHERE name = $1 AND channel = $2 AND locale = $3`,
		t.Name, t.Channel, t.Locale,
	)
	if err != nil {
		return err
	}

	// Insert new version
	t.Version++
	return e.CreateTemplate(ctx, t)
}

// InvalidateCache clears the template cache.
func (e *Engine) InvalidateCache() {
	e.mu.Lock()
	e.cache = make(map[string]*compiledTemplate)
	e.mu.Unlock()
}
