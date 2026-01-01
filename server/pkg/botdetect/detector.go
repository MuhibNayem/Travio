package botdetect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// DeviceFingerprint represents collected device signals
type DeviceFingerprint struct {
	// From HTTP headers
	UserAgent       string `json:"user_agent"`
	AcceptLanguage  string `json:"accept_language"`
	AcceptEncoding  string `json:"accept_encoding"`
	DNT             string `json:"dnt"`
	SecChUa         string `json:"sec_ch_ua"` // Client hints
	SecChUaPlatform string `json:"sec_ch_ua_platform"`
	SecChUaMobile   string `json:"sec_ch_ua_mobile"`

	// From client-side JS (sent in request body)
	ScreenWidth    int    `json:"screen_width,omitempty"`
	ScreenHeight   int    `json:"screen_height,omitempty"`
	ColorDepth     int    `json:"color_depth,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
	Language       string `json:"language,omitempty"`
	Platform       string `json:"platform,omitempty"`
	CookiesEnabled bool   `json:"cookies_enabled,omitempty"`
	JavaEnabled    bool   `json:"java_enabled,omitempty"`
	CanvasHash     string `json:"canvas_hash,omitempty"` // Canvas fingerprint
	WebGLVendor    string `json:"webgl_vendor,omitempty"`
	WebGLRenderer  string `json:"webgl_renderer,omitempty"`
	AudioHash      string `json:"audio_hash,omitempty"` // AudioContext fingerprint
	Fonts          string `json:"fonts,omitempty"`      // Available fonts hash
}

// RiskSignals contains bot detection analysis
type RiskSignals struct {
	FingerprintHash    string   `json:"fingerprint_hash"`
	RiskScore          float64  `json:"risk_score"` // 0.0 (safe) to 1.0 (bot)
	IsKnownBot         bool     `json:"is_known_bot"`
	IsDataCenterIP     bool     `json:"is_datacenter_ip"`
	IsHeadlessBrowser  bool     `json:"is_headless_browser"`
	HasInconsistencies bool     `json:"has_inconsistencies"`
	SeenCount          int64    `json:"seen_count"`           // How many times this fingerprint was seen
	UniqueIPsForPrint  int64    `json:"unique_ips_for_print"` // Different IPs using same fingerprint
	Reasons            []string `json:"reasons"`
}

// Detector performs bot detection
type Detector struct {
	client         *redis.Client
	knownBotUAs    []string
	datacenterASNs map[string]bool
}

// NewDetector creates a new bot detector
func NewDetector(client *redis.Client) *Detector {
	return &Detector{
		client: client,
		knownBotUAs: []string{
			"bot", "crawler", "spider", "scraper", "curl", "wget",
			"python-requests", "http-client", "java/", "go-http-client",
			"phantomjs", "headlesschrome", "puppeteer",
		},
		datacenterASNs: map[string]bool{
			// Common cloud providers (simplified)
			"AS14618": true, // Amazon
			"AS15169": true, // Google
			"AS8075":  true, // Microsoft
			"AS13335": true, // Cloudflare
			"AS16509": true, // Amazon
		},
	}
}

// ExtractFingerprint builds fingerprint from HTTP request
func (d *Detector) ExtractFingerprint(r *http.Request) *DeviceFingerprint {
	fp := &DeviceFingerprint{
		UserAgent:       r.Header.Get("User-Agent"),
		AcceptLanguage:  r.Header.Get("Accept-Language"),
		AcceptEncoding:  r.Header.Get("Accept-Encoding"),
		DNT:             r.Header.Get("DNT"),
		SecChUa:         r.Header.Get("Sec-CH-UA"),
		SecChUaPlatform: r.Header.Get("Sec-CH-UA-Platform"),
		SecChUaMobile:   r.Header.Get("Sec-CH-UA-Mobile"),
	}

	// Try to parse client-side fingerprint from X-Device-Fingerprint header (JSON)
	clientFP := r.Header.Get("X-Device-Fingerprint")
	if clientFP != "" {
		var clientData DeviceFingerprint
		if err := json.Unmarshal([]byte(clientFP), &clientData); err == nil {
			fp.ScreenWidth = clientData.ScreenWidth
			fp.ScreenHeight = clientData.ScreenHeight
			fp.ColorDepth = clientData.ColorDepth
			fp.Timezone = clientData.Timezone
			fp.Language = clientData.Language
			fp.Platform = clientData.Platform
			fp.CookiesEnabled = clientData.CookiesEnabled
			fp.CanvasHash = clientData.CanvasHash
			fp.WebGLVendor = clientData.WebGLVendor
			fp.WebGLRenderer = clientData.WebGLRenderer
			fp.AudioHash = clientData.AudioHash
			fp.Fonts = clientData.Fonts
		}
	}

	return fp
}

// HashFingerprint creates a stable hash of the fingerprint
func (d *Detector) HashFingerprint(fp *DeviceFingerprint) string {
	data, _ := json.Marshal(fp)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Analyze performs bot detection analysis
func (d *Detector) Analyze(ctx context.Context, fp *DeviceFingerprint, ip string) (*RiskSignals, error) {
	signals := &RiskSignals{
		FingerprintHash: d.HashFingerprint(fp),
		RiskScore:       0.0,
		Reasons:         []string{},
	}

	// Check for known bot user agents
	uaLower := strings.ToLower(fp.UserAgent)
	for _, botUA := range d.knownBotUAs {
		if strings.Contains(uaLower, botUA) {
			signals.IsKnownBot = true
			signals.RiskScore += 0.9
			signals.Reasons = append(signals.Reasons, "known_bot_user_agent")
			break
		}
	}

	// Check for headless browser indicators
	if d.isHeadlessBrowser(fp) {
		signals.IsHeadlessBrowser = true
		signals.RiskScore += 0.7
		signals.Reasons = append(signals.Reasons, "headless_browser_detected")
	}

	// Check for header inconsistencies
	if d.hasInconsistencies(fp) {
		signals.HasInconsistencies = true
		signals.RiskScore += 0.3
		signals.Reasons = append(signals.Reasons, "header_inconsistencies")
	}

	// Track fingerprint usage in Redis
	fpKey := fmt.Sprintf("botdetect:fp:%s", signals.FingerprintHash)
	ipSetKey := fmt.Sprintf("botdetect:fp_ips:%s", signals.FingerprintHash)

	// Increment seen count
	seenCount, _ := d.client.Incr(ctx, fpKey).Result()
	signals.SeenCount = seenCount
	d.client.Expire(ctx, fpKey, 24*time.Hour)

	// Track unique IPs for this fingerprint
	d.client.SAdd(ctx, ipSetKey, ip)
	d.client.Expire(ctx, ipSetKey, 24*time.Hour)
	uniqueIPs, _ := d.client.SCard(ctx, ipSetKey).Result()
	signals.UniqueIPsForPrint = uniqueIPs

	// Suspicious: Same fingerprint from many different IPs (distributed attack)
	if uniqueIPs > 10 {
		signals.RiskScore += 0.5
		signals.Reasons = append(signals.Reasons, "fingerprint_used_from_many_ips")
	}

	// Very high request count is suspicious
	if seenCount > 1000 {
		signals.RiskScore += 0.3
		signals.Reasons = append(signals.Reasons, "high_request_volume")
	}

	// Cap at 1.0
	if signals.RiskScore > 1.0 {
		signals.RiskScore = 1.0
	}

	return signals, nil
}

func (d *Detector) isHeadlessBrowser(fp *DeviceFingerprint) bool {
	ua := strings.ToLower(fp.UserAgent)

	// HeadlessChrome detection
	if strings.Contains(ua, "headlesschrome") {
		return true
	}

	// PhantomJS
	if strings.Contains(ua, "phantomjs") {
		return true
	}

	// Missing WebGL (common in headless)
	if fp.WebGLVendor == "" && fp.WebGLRenderer == "" && fp.CanvasHash != "" {
		return true
	}

	// Chrome but no client hints (possible automation)
	if strings.Contains(ua, "chrome") && fp.SecChUa == "" {
		return true
	}

	return false
}

func (d *Detector) hasInconsistencies(fp *DeviceFingerprint) bool {
	ua := strings.ToLower(fp.UserAgent)
	platform := strings.ToLower(fp.Platform)

	// Mobile UA but desktop platform
	if strings.Contains(ua, "mobile") && (platform == "win32" || platform == "linux x86_64" || platform == "macintel") {
		return true
	}

	// Windows UA but Mac platform
	if strings.Contains(ua, "windows") && strings.Contains(platform, "mac") {
		return true
	}

	// Linux UA but Windows platform
	if strings.Contains(ua, "linux") && strings.Contains(platform, "win") {
		return true
	}

	return false
}

// Middleware creates HTTP middleware for bot detection
func (d *Detector) Middleware(blockThreshold float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fp := d.ExtractFingerprint(r)
			signals, err := d.Analyze(r.Context(), fp, r.RemoteAddr)
			if err != nil {
				// On error, allow through
				next.ServeHTTP(w, r)
				return
			}

			// Set response headers for debugging/monitoring
			w.Header().Set("X-Bot-Score", fmt.Sprintf("%.2f", signals.RiskScore))
			w.Header().Set("X-Fingerprint", signals.FingerprintHash[:16])

			if signals.RiskScore >= blockThreshold {
				w.Header().Set("X-Block-Reason", strings.Join(signals.Reasons, ","))
				http.Error(w, "Request blocked", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
