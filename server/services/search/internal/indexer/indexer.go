package indexer

import (
	"context"
	"fmt"
	"strings"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type Indexer struct {
	client *opensearch.Client
}

func New(client *opensearch.Client) *Indexer {
	return &Indexer{client: client}
}

func (i *Indexer) InitIndices(ctx context.Context) error {
	indices := []string{"trips", "stations"}

	for _, idx := range indices {
		// Check if index exists
		req := opensearchapi.IndicesExistsRequest{
			Index: []string{idx},
		}
		res, err := req.Do(ctx, i.client)
		if err != nil {
			return err
		}
		if res.StatusCode == 200 {
			logger.Info("Index exists", "index", idx)
			continue
		}

		// Create index
		createReq := opensearchapi.IndicesCreateRequest{
			Index: idx,
		}
		res, err = createReq.Do(ctx, i.client)
		if err != nil {
			return err
		}
		if res.IsError() {
			logger.Error("Failed to create index", "index", idx, "status", res.Status())
		} else {
			logger.Info("Created index", "index", idx)
		}
	}
	return nil
}

func (i *Indexer) IndexDocument(ctx context.Context, index, id, body string) error {
	req := opensearchapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       strings.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, i.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		logger.Error("Error indexing document", "index", index, "id", id, "status", res.Status())
		return fmt.Errorf("failed to index document: %s", res.Status())
	}

	return nil
}
