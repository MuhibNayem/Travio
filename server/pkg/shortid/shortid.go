package shortid

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// GenerateShortCode creates a human-readable short code for user-facing IDs
// Format: PREFIX-YYYY-MMDD-XXXX (e.g., ORD-2026-0407-3F8B)
func GenerateShortCode(prefix string) string {
	now := time.Now()
	date := now.Format("2006-0102") // YYYY-MMDD
	suffix := randomHex(4)
	return fmt.Sprintf("%s-%s-%s", prefix, date, suffix)
}

// GenerateSimpleCode creates a shorter code without date
// Format: PREFIX-XXXX (e.g., TKT-3F8B)
func GenerateSimpleCode(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, randomHex(4))
}

// ValidateShortCode checks if a short code is well-formed
func ValidateShortCode(code string) bool {
	parts := strings.Split(code, "-")
	if len(parts) < 2 {
		return false
	}
	// Last part should be hex
	hex := parts[len(parts)-1]
	if len(hex) != 4 {
		return false
	}
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// ExtractPrefix returns the prefix from a short code
func ExtractPrefix(code string) string {
	parts := strings.Split(code, "-")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// IsShortCode returns true if the given string looks like a short code (not a UUID)
func IsShortCode(s string) bool {
	// UUID format: 8-4-4-4-12 with dashes
	if strings.Count(s, "-") == 4 {
		return false
	}
	// Short code: PREFIX-DATE-HEX or PREFIX-HEX
	return strings.Contains(s, "-") && len(s) < 30
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Fallback (should never happen)
		return fmt.Sprintf("%04X", time.Now().UnixNano()%65536)
	}
	return fmt.Sprintf("%X", b)[:n]
}
