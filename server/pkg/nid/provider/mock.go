package provider

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/nid"
)

// MockProvider implements nid.Provider for testing and development
// Returns deterministic responses based on NID patterns
type MockProvider struct {
	validator *nid.BangladeshNIDValidator
	delay     time.Duration // Simulate network latency
}

func NewMockProvider(delay time.Duration) *MockProvider {
	return &MockProvider{
		validator: &nid.BangladeshNIDValidator{},
		delay:     delay,
	}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) Country() string {
	return "BD"
}

func (p *MockProvider) Verify(ctx context.Context, req *nid.VerifyRequest) (*nid.VerifyResponse, error) {
	// Simulate network delay
	if p.delay > 0 {
		select {
		case <-time.After(p.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Validate format
	if err := p.validator.Validate(req.NID); err != nil {
		return &nid.VerifyResponse{
			IsValid:      false,
			ProviderName: p.Name(),
			VerifiedAt:   time.Now(),
			ErrorCode:    nid.ErrorCodeInvalidFormat,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Mock responses based on NID patterns
	nidStr := req.NID

	// Pattern: NIDs starting with "0" are "not found"
	if strings.HasPrefix(nidStr, "0") {
		return &nid.VerifyResponse{
			IsValid:      false,
			ProviderName: p.Name(),
			VerifiedAt:   time.Now(),
			ErrorCode:    nid.ErrorCodeNotFound,
			ErrorMessage: "NID not found in database",
		}, nil
	}

	// Pattern: NIDs starting with "9" have DOB mismatch
	if strings.HasPrefix(nidStr, "9") && !req.DateOfBirth.IsZero() {
		return &nid.VerifyResponse{
			IsValid:      false,
			ProviderName: p.Name(),
			VerifiedAt:   time.Now(),
			ErrorCode:    nid.ErrorCodeDOBMismatch,
			ErrorMessage: "date of birth does not match records",
		}, nil
	}

	// Success case: Generate mock citizen data
	citizen := p.generateMockCitizen(nidStr, req.DateOfBirth)

	// Calculate match score if name was provided
	var matchScore *nid.MatchScore
	if req.Name != "" {
		matchScore = &nid.MatchScore{
			NameMatch:    p.calculateNameMatch(req.Name, citizen.NameEN),
			DOBMatch:     true,
			OverallMatch: 0.95,
		}
	}

	return &nid.VerifyResponse{
		IsValid:      true,
		Confidence:   0.99,
		MatchScore:   matchScore,
		Citizen:      citizen,
		ProviderName: p.Name(),
		VerifiedAt:   time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}

func (p *MockProvider) HealthCheck(ctx context.Context) error {
	return nil // Mock is always healthy
}

func (p *MockProvider) generateMockCitizen(nidStr string, dob time.Time) *nid.CitizenData {
	// Generate deterministic mock data based on NID
	seed := int64(0)
	for _, c := range nidStr {
		seed = seed*10 + int64(c-'0')
	}
	r := rand.New(rand.NewSource(seed))

	firstNames := []string{"Mohammad", "Abdul", "Fatima", "Ayesha", "Rahim", "Karim"}
	lastNames := []string{"Rahman", "Islam", "Hossain", "Ahmed", "Khan", "Begum"}

	firstName := firstNames[r.Intn(len(firstNames))]
	lastName := lastNames[r.Intn(len(lastNames))]

	gender := "male"
	if r.Float32() > 0.5 {
		gender = "female"
	}

	if dob.IsZero() {
		// Generate a DOB between 18 and 80 years ago
		years := r.Intn(62) + 18
		dob = time.Now().AddDate(-years, -r.Intn(12), -r.Intn(28))
	}

	divisions := []string{"Dhaka", "Chittagong", "Rajshahi", "Khulna", "Sylhet", "Rangpur", "Barishal", "Mymensingh"}
	districts := []string{"Dhaka", "Gazipur", "Narayanganj", "Tangail", "Manikganj", "Munshiganj"}

	return &nid.CitizenData{
		NID:         nidStr,
		NameBN:      "মোঃ " + firstName + " " + lastName, // Mock Bengali
		NameEN:      firstName + " " + lastName,
		FatherName:  firstNames[r.Intn(len(firstNames))] + " " + lastNames[r.Intn(len(lastNames))],
		MotherName:  firstNames[r.Intn(len(firstNames))] + " " + lastNames[r.Intn(len(lastNames))],
		DateOfBirth: dob,
		Gender:      gender,
		BloodGroup:  []string{"A+", "A-", "B+", "B-", "O+", "O-", "AB+", "AB-"}[r.Intn(8)],
		PresentAddress: &nid.Address{
			Division: divisions[r.Intn(len(divisions))],
			District: districts[r.Intn(len(districts))],
			Upazila:  "Mock Upazila",
			PostCode: "1000",
		},
		PermanentAddress: &nid.Address{
			Division: divisions[r.Intn(len(divisions))],
			District: districts[r.Intn(len(districts))],
			Upazila:  "Mock Upazila",
			PostCode: "1000",
		},
		VoterArea: "Ward " + string(rune('1'+r.Intn(9))),
	}
}

func (p *MockProvider) calculateNameMatch(provided, actual string) float64 {
	// Simple Levenshtein-based similarity
	provided = strings.ToLower(strings.TrimSpace(provided))
	actual = strings.ToLower(strings.TrimSpace(actual))

	if provided == actual {
		return 1.0
	}

	// Check if one contains the other
	if strings.Contains(actual, provided) || strings.Contains(provided, actual) {
		return 0.8
	}

	// Calculate word overlap
	providedWords := strings.Fields(provided)
	actualWords := strings.Fields(actual)

	matches := 0
	for _, pw := range providedWords {
		for _, aw := range actualWords {
			if pw == aw {
				matches++
				break
			}
		}
	}

	if len(providedWords) == 0 {
		return 0
	}

	return float64(matches) / float64(max(len(providedWords), len(actualWords)))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
