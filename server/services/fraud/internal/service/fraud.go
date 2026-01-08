// Package service implements fraud detection logic.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/client"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/profile"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/rag"
)

const (
	// Default threshold for blocking transactions
	DefaultBlockThreshold = 70
)

// FraudService handles fraud detection operations.
type FraudService struct {
	zaiClient      *client.Client
	blockThreshold int

	// User profiling
	profileStore    *profile.Store
	profileAnalyzer *profile.Analyzer

	// RAG
	embedder  *rag.Embedder
	retriever *rag.Retriever
}

// NewFraudService creates a new fraud service.
func NewFraudService(zaiClient *client.Client, blockThreshold int) *FraudService {
	if blockThreshold <= 0 {
		blockThreshold = DefaultBlockThreshold
	}
	return &FraudService{
		zaiClient:      zaiClient,
		blockThreshold: blockThreshold,
	}
}

// WithProfiling adds user profiling capabilities.
func (s *FraudService) WithProfiling(store *profile.Store, analyzer *profile.Analyzer) *FraudService {
	s.profileStore = store
	s.profileAnalyzer = analyzer
	return s
}

// WithRAG adds RAG capabilities.
func (s *FraudService) WithRAG(embedder *rag.Embedder, retriever *rag.Retriever) *FraudService {
	s.embedder = embedder
	s.retriever = retriever
	return s
}

// AnalyzeBooking performs behavioral fraud analysis on a booking.
func (s *FraudService) AnalyzeBooking(ctx context.Context, req *domain.BookingAnalysisRequest) (*domain.FraudResult, error) {
	// Generate cache key
	cacheKey := s.generateCacheKey("booking", req.OrderID, req.UserID)

	// Check cache
	if cached, err := s.zaiClient.GetCached(ctx, cacheKey); err == nil && cached != "" {
		var result domain.FraudResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			logger.Debug("Fraud analysis cache hit", "order_id", req.OrderID)
			return &result, nil
		}
	}

	// === User Profile Analysis ===
	var deviationScore float64
	var profileContext string
	if s.profileStore != nil && s.profileAnalyzer != nil {
		userProfile, _ := s.profileStore.GetProfile(ctx, req.UserID)
		event := &profile.BookingEvent{
			UserID:      req.UserID,
			OrderID:     req.OrderID,
			TripID:      req.TripID,
			AmountPaisa: req.TotalAmountPaisa,
			BookingTime: req.BookingTimestamp,
			IPAddress:   req.IPAddress,
			UserAgent:   req.UserAgent,
		}
		deviation := s.profileAnalyzer.AnalyzeDeviation(userProfile, event)
		deviationScore = deviation.Score

		if deviation.IsNewUser {
			profileContext = "New user - limited history available."
		} else if deviation.IsAnomalous {
			profileContext = fmt.Sprintf("ANOMALOUS behavior detected (deviation score: %.0f). Flags: ", deviation.Score)
			if deviation.VelocityDeviation > 2.0 {
				profileContext += "high_velocity "
			}
			if deviation.ValueDeviation > 2.0 {
				profileContext += "unusual_value "
			}
			if deviation.IPDeviation > 0 {
				profileContext += "new_ip "
			}
			if deviation.DeviceDeviation > 0 {
				profileContext += "new_device "
			}
		} else {
			profileContext = "User behavior consistent with historical patterns."
		}
	}

	// === RAG: Retrieve Similar Cases ===
	var ragContext string
	if s.embedder != nil && s.retriever != nil {
		route := ""
		if len(req.PassengerNIDs) > 0 {
			route = req.TripID // Use trip as route identifier
		}
		embedding, err := s.embedder.EmbedBooking(ctx, req.OrderID, req.UserID, route,
			req.TotalAmountPaisa, req.PassengerCount, req.IPAddress, req.UserAgent, 0)
		if err == nil && len(embedding) > 0 {
			similarCases, _ := s.retriever.RetrieveSimilar(ctx, embedding)
			if len(similarCases) > 0 {
				context := s.retriever.BuildContext(similarCases)
				ragContext = context.Summary
			}
		}
	}

	// Build enhanced analysis prompt
	systemPrompt := `You are a fraud detection AI for a travel booking system. Analyze booking data for fraud indicators.
Respond ONLY with valid JSON in this exact format:
{
  "risk_score": <0-100>,
  "confidence": <0-100>,
  "risk_factors": [
    {"code": "<CODE>", "description": "<desc>", "severity": "<low|medium|high>", "score": <0-30>}
  ],
  "summary": "<one sentence summary>"
}

Common fraud indicators:
- Velocity abuse: Too many bookings in short time
- Identity mismatch: Same NID with different names
- Suspicious timing: Bookings at unusual hours
- High-value unusual patterns
- Known problematic IP ranges
- Bot-like behavior patterns`

	userPrompt := fmt.Sprintf(`Analyze this booking for fraud:

Order ID: %s
User ID: %s
Trip ID: %s
Passenger Count: %d
Passenger NIDs: %s
Passenger Names: %s
Booking Time: %s
IP Address: %s
User Agent: %s
Payment Method: %s
Total Amount: %d paisa
Bookings Last 24h: %d
Bookings Last Week: %d
Previous Fraud Flags: %d`,
		req.OrderID, req.UserID, req.TripID, req.PassengerCount,
		strings.Join(req.PassengerNIDs, ", "),
		strings.Join(req.PassengerNames, ", "),
		req.BookingTimestamp.Format(time.RFC3339),
		req.IPAddress, req.UserAgent, req.PaymentMethod,
		req.TotalAmountPaisa, req.BookingsLast24Hours,
		req.BookingsLastWeek, req.PreviousFraudFlags,
	)

	// Add profile context if available
	if profileContext != "" {
		userPrompt += fmt.Sprintf("\n\n--- USER PROFILE ANALYSIS ---\n%s", profileContext)
	}

	// Add RAG context if available
	if ragContext != "" {
		userPrompt += fmt.Sprintf("\n\n--- SIMILAR HISTORICAL CASES ---\n%s", ragContext)
	}

	userPrompt += "\n\nProvide fraud risk assessment."

	// Call Z.AI
	response, err := s.zaiClient.AnalyzeText(ctx, systemPrompt, userPrompt)
	if err != nil {
		logger.Error("Z.AI fraud analysis failed", "order_id", req.OrderID, "error", err)
		// Fail-open: return risk based on deviation score if AI unavailable
		if deviationScore > 50 {
			return s.defaultRiskResult(int(deviationScore), "AI unavailable, using profile deviation"), nil
		}
		return s.defaultLowRiskResult("AI analysis unavailable"), nil
	}

	// Parse response
	result, err := s.parseAnalysisResponse(response)
	if err != nil {
		logger.Error("Failed to parse Z.AI response", "order_id", req.OrderID, "error", err, "response", response)
		return s.defaultLowRiskResult("Failed to parse analysis"), nil
	}

	// Adjust score based on profile deviation (max +20 from profile)
	if deviationScore > 30 {
		adjustment := int(deviationScore * 0.2) // 20% of deviation score
		result.RiskScore = min(100, result.RiskScore+adjustment)
	}

	result.Model = client.ModelGLM45Flash
	result.AnalyzedAt = time.Now()
	result.RiskLevel = domain.RiskScoreToLevel(result.RiskScore)
	result.ShouldBlock = result.RiskScore >= s.blockThreshold

	// Cache result
	if cacheData, err := json.Marshal(result); err == nil {
		_ = s.zaiClient.SetCached(ctx, cacheKey, string(cacheData))
	}

	// Update user profile asynchronously
	if s.profileStore != nil {
		go func() {
			event := &profile.BookingEvent{
				UserID:      req.UserID,
				OrderID:     req.OrderID,
				TripID:      req.TripID,
				AmountPaisa: req.TotalAmountPaisa,
				BookingTime: req.BookingTimestamp,
				IPAddress:   req.IPAddress,
				UserAgent:   req.UserAgent,
				RiskScore:   float64(result.RiskScore),
				WasBlocked:  result.ShouldBlock,
			}
			if err := s.profileStore.UpdateFromEvent(context.Background(), event); err != nil {
				logger.Warn("Failed to update user profile", "user_id", req.UserID, "error", err)
			}
		}()
	}

	logger.Info("Fraud analysis completed",
		"order_id", req.OrderID,
		"risk_score", result.RiskScore,
		"risk_level", result.RiskLevel,
		"should_block", result.ShouldBlock,
		"deviation_score", deviationScore,
		"has_rag_context", ragContext != "",
	)

	return result, nil
}

// VerifyDocument performs document verification using vision model.
func (s *FraudService) VerifyDocument(ctx context.Context, req *domain.DocumentVerificationRequest) (*domain.DocumentVerificationResult, error) {
	// Generate cache key from image hash
	imageHash := sha256.Sum256(req.DocumentImage)
	cacheKey := s.generateCacheKey("doc", hex.EncodeToString(imageHash[:8]))

	// Check cache
	if cached, err := s.zaiClient.GetCached(ctx, cacheKey); err == nil && cached != "" {
		var result domain.DocumentVerificationResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			logger.Debug("Document verification cache hit")
			return &result, nil
		}
	}

	systemPrompt := `You are a document verification AI. Analyze the provided ID document image.
Respond ONLY with valid JSON in this exact format:
{
  "is_authentic": <true|false>,
  "confidence": <0-100>,
  "extracted_nid": "<NID number if visible>",
  "extracted_name": "<Name if visible>",
  "tampering_score": <0-100>,
  "issues": ["<list of issues found>"],
  "summary": "<one sentence summary>"
}

Check for:
- Image tampering or editing artifacts
- Text consistency and alignment
- Proper document format
- Visible security features
- Photo quality and authenticity`

	userPrompt := fmt.Sprintf(`Verify this %s document.
Expected NID: %s
Expected Name: %s

Analyze the image for authenticity and tampering.`,
		req.DocumentType, req.ExpectedNID, req.ExpectedName)

	// Call Z.AI vision model
	response, err := s.zaiClient.AnalyzeImage(ctx, systemPrompt, userPrompt, req.DocumentImage, req.ImageMimeType)
	if err != nil {
		logger.Error("Z.AI document verification failed", "error", err)
		return &domain.DocumentVerificationResult{
			IsAuthentic: false,
			Confidence:  0,
			Summary:     "Verification unavailable: " + err.Error(),
		}, nil
	}

	// Parse response
	result, err := s.parseDocumentResponse(response)
	if err != nil {
		logger.Error("Failed to parse document verification response", "error", err, "response", response)
		return &domain.DocumentVerificationResult{
			IsAuthentic: false,
			Confidence:  0,
			Summary:     "Failed to parse verification result",
		}, nil
	}

	// Cache result
	if cacheData, err := json.Marshal(result); err == nil {
		_ = s.zaiClient.SetCached(ctx, cacheKey, string(cacheData))
	}

	logger.Info("Document verification completed",
		"is_authentic", result.IsAuthentic,
		"confidence", result.Confidence,
		"tampering_score", result.TamperingScore,
	)

	return result, nil
}

// parseAnalysisResponse parses the Z.AI response into FraudResult.
func (s *FraudService) parseAnalysisResponse(response string) (*domain.FraudResult, error) {
	jsonStr := extractJSON(response)

	var parsed struct {
		RiskScore   int                 `json:"risk_score"`
		Confidence  int                 `json:"confidence"`
		RiskFactors []domain.RiskFactor `json:"risk_factors"`
		Summary     string              `json:"summary"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &domain.FraudResult{
		RiskScore:   parsed.RiskScore,
		Confidence:  parsed.Confidence,
		RiskFactors: parsed.RiskFactors,
		Summary:     parsed.Summary,
	}, nil
}

// parseDocumentResponse parses the Z.AI document verification response.
func (s *FraudService) parseDocumentResponse(response string) (*domain.DocumentVerificationResult, error) {
	jsonStr := extractJSON(response)

	var result domain.DocumentVerificationResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// defaultLowRiskResult returns a default low-risk result for fail-open scenarios.
func (s *FraudService) defaultLowRiskResult(reason string) *domain.FraudResult {
	return s.defaultRiskResult(0, reason)
}

// defaultRiskResult returns a default result with specified score.
func (s *FraudService) defaultRiskResult(score int, reason string) *domain.FraudResult {
	return &domain.FraudResult{
		RiskScore:   score,
		RiskLevel:   domain.RiskScoreToLevel(score),
		Confidence:  0,
		ShouldBlock: score >= s.blockThreshold,
		Summary:     reason,
		AnalyzedAt:  time.Now(),
		Model:       "default",
	}
}

// generateCacheKey generates a cache key for fraud analysis.
func (s *FraudService) generateCacheKey(prefix string, parts ...string) string {
	key := "fraud:" + prefix
	for _, p := range parts {
		key += ":" + p
	}
	return key
}

// extractJSON extracts JSON from a string that might be wrapped in markdown.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	}

	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		s = s[start : end+1]
	}

	return strings.TrimSpace(s)
}
