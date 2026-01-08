// Package client provides an HTTP client for Z.AI GLM API.
package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	// Model names
	ModelGLM45Flash  = "glm-4.5-flash"  // Text analysis
	ModelGLM46VFlash = "glm-4.6v-flash" // Vision analysis

	// Default configuration
	DefaultBaseURL    = "https://api.z.ai/api/paas/v4"
	DefaultTimeout    = 30 * time.Second
	DefaultCacheTTL   = 5 * time.Minute
	DefaultMaxRetries = 3
)

// Config holds Z.AI client configuration.
type Config struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	CacheTTL   time.Duration
	MaxRetries int
}

// LoadConfigFromEnv loads configuration from environment variables.
func LoadConfigFromEnv() Config {
	baseURL := os.Getenv("ZAI_BASE_URL")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	timeout := DefaultTimeout
	if t := os.Getenv("ZAI_TIMEOUT"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			timeout = d
		}
	}

	cacheTTL := DefaultCacheTTL
	if t := os.Getenv("FRAUD_CACHE_TTL"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			cacheTTL = d
		}
	}

	return Config{
		APIKey:     os.Getenv("ZAI_API_KEY"),
		BaseURL:    baseURL,
		Timeout:    timeout,
		CacheTTL:   cacheTTL,
		MaxRetries: DefaultMaxRetries,
	}
}

// Client is the Z.AI API client.
type Client struct {
	httpClient *http.Client
	config     Config
	cache      *redis.Client
}

// NewClient creates a new Z.AI client.
func NewClient(cfg Config, redisClient *redis.Client) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
		cache:  redisClient,
	}
}

// Message represents a chat message.
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []ContentPart
}

// ContentPart represents a part of multimodal content.
type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL in content.
type ImageURL struct {
	URL string `json:"url"` // Can be URL or base64 data URI
}

// ChatRequest is the request body for chat completions.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse is the response from chat completions.
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletion sends a chat completion request.
func (c *Client) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if c.config.APIKey == "" {
		return nil, fmt.Errorf("ZAI_API_KEY not configured")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.config.BaseURL)

	var resp *ChatResponse
	var lastErr error

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		httpReq.Header.Set("Accept-Language", "en-US,en")

		httpResp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if httpResp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API returned status %d: %s", httpResp.StatusCode, string(respBody))
			if httpResp.StatusCode >= 500 {
				continue // Retry on server errors
			}
			return nil, lastErr // Don't retry on client errors
		}

		if err := json.Unmarshal(respBody, &resp); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

// AnalyzeText sends a text analysis request using GLM-4.5-Flash.
func (c *Client) AnalyzeText(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	req := &ChatRequest{
		Model: ModelGLM45Flash,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.1, // Low temperature for consistent analysis
		MaxTokens:   1024,
	}

	resp, err := c.ChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content, ok := resp.Choices[0].Message.Content.(string)
	if !ok {
		return "", fmt.Errorf("unexpected content type in response")
	}

	logger.Debug("Z.AI text analysis completed",
		"model", ModelGLM45Flash,
		"tokens", resp.Usage.TotalTokens,
	)

	return content, nil
}

// AnalyzeImage sends an image analysis request using GLM-4.6V-Flash.
func (c *Client) AnalyzeImage(ctx context.Context, systemPrompt, userPrompt string, imageData []byte, mimeType string) (string, error) {
	// Convert image to base64 data URI
	b64 := base64.StdEncoding.EncodeToString(imageData)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, b64)

	// Build multimodal content
	content := []ContentPart{
		{Type: "text", Text: userPrompt},
		{Type: "image_url", ImageURL: &ImageURL{URL: dataURI}},
	}

	req := &ChatRequest{
		Model: ModelGLM46VFlash,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: content},
		},
		Temperature: 0.1,
		MaxTokens:   1024,
	}

	resp, err := c.ChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	result, ok := resp.Choices[0].Message.Content.(string)
	if !ok {
		return "", fmt.Errorf("unexpected content type in response")
	}

	logger.Debug("Z.AI image analysis completed",
		"model", ModelGLM46VFlash,
		"tokens", resp.Usage.TotalTokens,
	)

	return result, nil
}

// CacheKey generates a cache key for a request.
func (c *Client) CacheKey(prefix string, data ...string) string {
	key := "fraud:" + prefix
	for _, d := range data {
		key += ":" + d
	}
	return key
}

// GetCached retrieves a cached result.
func (c *Client) GetCached(ctx context.Context, key string) (string, error) {
	if c.cache == nil {
		return "", nil
	}
	result, err := c.cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

// SetCached stores a result in cache.
func (c *Client) SetCached(ctx context.Context, key, value string) error {
	if c.cache == nil {
		return nil
	}
	return c.cache.Set(ctx, key, value, c.config.CacheTTL).Err()
}
