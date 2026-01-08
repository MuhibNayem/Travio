package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

const (
	// Google Gemini embedding model (free tier)
	EmbeddingModel = "text-embedding-005"
	EmbeddingDims  = 768 // text-embedding-005 produces 768-dim vectors

	// Google Gemini API endpoint
	GeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// Embedder generates text embeddings using Google Gemini API.
type Embedder struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewEmbedder creates a new embedder.
func NewEmbedder() *Embedder {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY") // Alternative env var
	}

	return &Embedder{
		apiKey:  apiKey,
		baseURL: GeminiBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GeminiEmbedRequest is the request body for Gemini embedding API.
type GeminiEmbedRequest struct {
	Model   string        `json:"model"`
	Content GeminiContent `json:"content"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiEmbedResponse is the response from Gemini embedding API.
type GeminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

// Embed generates an embedding for the given text using Google's text-embedding-005.
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if e.apiKey == "" {
		logger.Warn("GOOGLE_API_KEY not configured, returning zero embedding")
		return make([]float32, EmbeddingDims), nil
	}

	reqBody := GeminiEmbedRequest{
		Model: fmt.Sprintf("models/%s", EmbeddingModel),
		Content: GeminiContent{
			Parts: []GeminiPart{{Text: text}},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:embedContent?key=%s", e.baseURL, EmbeddingModel, e.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp GeminiEmbedResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResp.Embedding.Values) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	logger.Debug("Generated embedding",
		"model", EmbeddingModel,
		"dims", len(apiResp.Embedding.Values),
	)

	return apiResp.Embedding.Values, nil
}

// EmbedBooking generates an embedding for booking data.
func (e *Embedder) EmbedBooking(ctx context.Context, orderID, userID, route string, amount int64, passengerCount int, ipAddress, userAgent string, riskScore int) ([]float32, error) {
	text := BookingToText(orderID, userID, route, amount, passengerCount, ipAddress, userAgent, riskScore)
	return e.Embed(ctx, text)
}
