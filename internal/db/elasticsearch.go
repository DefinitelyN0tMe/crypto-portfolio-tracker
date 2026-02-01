package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticSearch struct {
	Client *elasticsearch.Client
}

func NewElasticSearch(addresses []string) (*ElasticSearch, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Check connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	log.Println("✅ Connected to ElasticSearch")
	return &ElasticSearch{Client: client}, nil
}

func (es *ElasticSearch) InitIndex() error {
	indexName := "crypto_tokens"

	// Check if index exists
	res, err := es.Client.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("failed to check index: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		log.Printf("✅ Index '%s' already exists", indexName)
		return nil
	}

	// Create index with mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":            map[string]interface{}{"type": "keyword"},
				"symbol":        map[string]interface{}{"type": "keyword"},
				"name":          map[string]interface{}{"type": "text"},
				"current_price": map[string]interface{}{"type": "double"},
				"market_cap":    map[string]interface{}{"type": "double"},
				"volume_24h":    map[string]interface{}{"type": "double"},
				"updated_at":    map[string]interface{}{"type": "date"},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("failed to encode mapping: %w", err)
	}

	res, err = es.Client.Indices.Create(
		indexName,
		es.Client.Indices.Create.WithBody(&buf),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	log.Printf("✅ Created index '%s'", indexName)
	return nil
}

func (es *ElasticSearch) IndexToken(ctx context.Context, token map[string]interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(token); err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	res, err := es.Client.Index(
		"crypto_tokens",
		&buf,
		es.Client.Index.WithDocumentID(token["id"].(string)),
		es.Client.Index.WithContext(ctx),
		es.Client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to index token: %w", err)
	}
	defer res.Body.Close()

	return nil
}

func (es *ElasticSearch) SearchTokens(ctx context.Context, query string) ([]map[string]interface{}, error) {
	var buf bytes.Buffer
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name", "symbol"},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("failed to encode search query: %w", err)
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("crypto_tokens"),
		es.Client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	tokens := make([]map[string]interface{}, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		tokens = append(tokens, source)
	}

	return tokens, nil
}
