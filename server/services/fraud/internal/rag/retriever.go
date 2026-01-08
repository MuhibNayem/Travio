package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"gorm.io/gorm"
)

const (
	// Index name for fraud cases
	FraudCasesIndex = "fraud_cases"
	// Default number of similar cases to retrieve
	DefaultTopK = 5
	// Embedding dimensions for text-embedding-005
	EmbeddingDimension = 768
)

// Retriever retrieves similar fraud cases using OpenSearch kNN.
type Retriever struct {
	osClient *opensearch.Client
	db       *gorm.DB
	embedder *Embedder
	topK     int
	minScore float64
}

// NewRetriever creates a new RAG retriever.
func NewRetriever(osClient *opensearch.Client, db *gorm.DB, embedder *Embedder, topK int, minScore float64) *Retriever {
	if topK <= 0 {
		topK = DefaultTopK
	}
	if minScore <= 0 {
		minScore = 0.7
	}
	return &Retriever{
		osClient: osClient,
		db:       db,
		embedder: embedder,
		topK:     topK,
		minScore: minScore,
	}
}

// InitializeIndex creates the OpenSearch index with kNN mapping.
func (r *Retriever) InitializeIndex(ctx context.Context) error {
	mapping := `{
		"settings": {
			"index": {
				"knn": true,
				"knn.algo_param.ef_search": 100
			}
		},
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"organization_id": {"type": "keyword"},
				"user_id": {"type": "keyword"},
				"order_id": {"type": "keyword"},
				"route": {"type": "keyword"},
				"risk_score": {"type": "integer"},
				"outcome": {"type": "keyword"},
				"created_at": {"type": "date"},
				"embedding": {
					"type": "knn_vector",
					"dimension": 768,
					"method": {
						"name": "hnsw",
						"space_type": "cosinesimil",
						"engine": "nmslib"
					}
				}
			}
		}
	}`

	req := opensearchapi.IndicesCreateRequest{
		Index: FraudCasesIndex,
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(ctx, r.osClient)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	// Ignore "already exists" error
	if res.IsError() && !strings.Contains(res.String(), "already_exists") {
		return fmt.Errorf("index creation failed: %s", res.String())
	}

	logger.Info("OpenSearch fraud_cases index ready")
	return nil
}

// IndexCase indexes a fraud case in OpenSearch.
func (r *Retriever) IndexCase(ctx context.Context, fraudCase *FraudCase) error {
	if len(fraudCase.Embedding) == 0 {
		return fmt.Errorf("case has no embedding")
	}

	doc, err := json.Marshal(fraudCase)
	if err != nil {
		return fmt.Errorf("failed to marshal case: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      FraudCasesIndex,
		DocumentID: fraudCase.ID,
		Body:       strings.NewReader(string(doc)),
	}

	res, err := req.Do(ctx, r.osClient)
	if err != nil {
		return fmt.Errorf("failed to index case: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing failed: %s", res.String())
	}

	return nil
}

// RetrieveSimilar retrieves similar fraud cases using kNN search.
func (r *Retriever) RetrieveSimilar(ctx context.Context, embedding []float32) ([]SimilarCase, error) {
	if r.osClient == nil || len(embedding) == 0 {
		return nil, nil
	}

	query := map[string]interface{}{
		"size": r.topK,
		"query": map[string]interface{}{
			"knn": map[string]interface{}{
				"embedding": map[string]interface{}{
					"vector": embedding,
					"k":      r.topK,
				},
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{FraudCasesIndex},
		Body:  strings.NewReader(string(body)),
	}

	res, err := req.Do(ctx, r.osClient)
	if err != nil {
		logger.Warn("OpenSearch search failed", "error", err)
		return nil, nil // Fail gracefully
	}
	defer res.Body.Close()

	if res.IsError() {
		logger.Warn("OpenSearch search error", "response", res.String())
		return nil, nil
	}

	var searchResult struct {
		Hits struct {
			Hits []struct {
				ID     string    `json:"_id"`
				Score  float64   `json:"_score"`
				Source FraudCase `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var cases []SimilarCase
	for _, hit := range searchResult.Hits.Hits {
		if hit.Score >= r.minScore {
			fraudCase := hit.Source
			fraudCase.ID = hit.ID
			cases = append(cases, SimilarCase{
				Case:       &fraudCase,
				Similarity: hit.Score,
			})
		}
	}

	logger.Debug("Retrieved similar cases", "count", len(cases))
	return cases, nil
}

// BuildContext builds RAG context from similar cases.
func (r *Retriever) BuildContext(cases []SimilarCase) *RAGContext {
	if len(cases) == 0 {
		return &RAGContext{
			Summary: "No similar historical cases found.",
		}
	}

	var summaryParts []string
	for i, c := range cases {
		outcome := c.Case.Outcome
		if outcome == "" {
			outcome = "unknown"
		}
		summaryParts = append(summaryParts, fmt.Sprintf(
			"Case %d (Outcome: %s, Similarity: %.0f%%): Route=%s, Risk=%d, Passengers=%d",
			i+1, outcome, c.Similarity*100,
			c.Case.Route, c.Case.RiskScore, c.Case.PassengerCount,
		))
	}

	return &RAGContext{
		Cases:   cases,
		Summary: strings.Join(summaryParts, "\n"),
	}
}

// SaveCase saves a fraud case to both PostgreSQL and OpenSearch.
func (r *Retriever) SaveCase(ctx context.Context, fraudCase *FraudCase) error {
	// Save to PostgreSQL
	if err := r.db.WithContext(ctx).Create(fraudCase).Error; err != nil {
		return fmt.Errorf("failed to save to DB: %w", err)
	}

	// Index in OpenSearch
	if r.osClient != nil && len(fraudCase.Embedding) > 0 {
		if err := r.IndexCase(ctx, fraudCase); err != nil {
			logger.Warn("Failed to index case in OpenSearch", "error", err)
			// Don't fail - DB save succeeded
		}
	}

	return nil
}
