package handlers

import (
	"bytes"
	"context"
	"crypto-portfolio-tracker/internal/db"
	"crypto-portfolio-tracker/internal/models"
	"crypto-portfolio-tracker/internal/services"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	ScyllaDB      *db.ScyllaDB
	ElasticSearch *db.ElasticSearch
}

func NewHandler(scylla *db.ScyllaDB, es *db.ElasticSearch) *Handler {
	return &Handler{
		ScyllaDB:      scylla,
		ElasticSearch: es,
	}
}

// Health check endpoint
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now(),
	})
}

// Add a token to both ScyllaDB and ElasticSearch
func (h *Handler) AddToken(c *fiber.Ctx) error {
	var token models.Token
	if err := c.BodyParser(&token); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	token.UpdatedAt = time.Now()

	// Insert into ScyllaDB
	query := `INSERT INTO tokens (id, symbol, name, current_price, market_cap, volume_24h, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`

	if err := h.ScyllaDB.Session.Query(query,
		token.ID, token.Symbol, token.Name, token.CurrentPrice,
		token.MarketCap, token.Volume24h, token.UpdatedAt).Exec(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to insert into ScyllaDB"})
	}

	// Index in ElasticSearch
	tokenMap := map[string]interface{}{
		"id":            token.ID,
		"symbol":        token.Symbol,
		"name":          token.Name,
		"current_price": token.CurrentPrice,
		"market_cap":    token.MarketCap,
		"volume_24h":    token.Volume24h,
		"updated_at":    token.UpdatedAt,
	}

	if err := h.ElasticSearch.IndexToken(context.Background(), tokenMap); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to index in ElasticSearch"})
	}

	return c.Status(201).JSON(token)
}

// Search tokens using ElasticSearch
func (h *Handler) SearchTokens(c *fiber.Ctx) error {
	query := c.Query("q", "")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Query parameter 'q' is required"})
	}

	tokens, err := h.ElasticSearch.SearchTokens(context.Background(), query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Search failed"})
	}

	return c.JSON(fiber.Map{
		"query":   query,
		"results": tokens,
		"count":   len(tokens),
	})
}

// Get token by ID from ScyllaDB
func (h *Handler) GetToken(c *fiber.Ctx) error {
	tokenID := c.Params("id")

	var token models.Token
	query := `SELECT id, symbol, name, current_price, market_cap, volume_24h, updated_at 
              FROM tokens WHERE id = ? LIMIT 1`

	if err := h.ScyllaDB.Session.Query(query, tokenID).Scan(
		&token.ID, &token.Symbol, &token.Name, &token.CurrentPrice,
		&token.MarketCap, &token.Volume24h, &token.UpdatedAt); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Token not found"})
	}

	return c.JSON(token)
}

// Sync tokens from CoinGecko
func (h *Handler) SyncTokens(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "10")
	limit := 10
	fmt.Sscanf(limitStr, "%d", &limit)

	// Create CoinGecko client
	cgClient := services.NewCoinGeckoClient()

	// Fetch tokens
	tokens, err := cgClient.FetchTopTokens(limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Save to ScyllaDB and ElasticSearch
	successCount := 0
	for _, token := range tokens {
		// Insert into ScyllaDB
		query := `INSERT INTO tokens (id, symbol, name, current_price, market_cap, volume_24h, updated_at) 
                  VALUES (?, ?, ?, ?, ?, ?, ?)`

		if err := h.ScyllaDB.Session.Query(query,
			token.ID, token.Symbol, token.Name, token.CurrentPrice,
			token.MarketCap, token.Volume24h, token.UpdatedAt).Exec(); err != nil {
			continue // Skip failed inserts
		}

		// Index in ElasticSearch
		tokenMap := map[string]interface{}{
			"id":            token.ID,
			"symbol":        token.Symbol,
			"name":          token.Name,
			"current_price": token.CurrentPrice,
			"market_cap":    token.MarketCap,
			"volume_24h":    token.Volume24h,
			"updated_at":    token.UpdatedAt,
		}

		if err := h.ElasticSearch.IndexToken(context.Background(), tokenMap); err != nil {
			continue
		}

		successCount++
	}

	return c.JSON(fiber.Map{
		"message": "Sync completed",
		"synced":  successCount,
		"total":   len(tokens),
	})
}

// Get price history for a token
func (h *Handler) GetPriceHistory(c *fiber.Ctx) error {
	tokenID := c.Params("id")
	limitStr := c.Query("limit", "100")
	limit := 100
	fmt.Sscanf(limitStr, "%d", &limit)

	query := `SELECT token_id, timestamp, price FROM price_history 
              WHERE token_id = ? LIMIT ?`

	iter := h.ScyllaDB.Session.Query(query, tokenID, limit).Iter()

	type PricePoint struct {
		TokenID   string    `json:"token_id"`
		Timestamp time.Time `json:"timestamp"`
		Price     float64   `json:"price"`
	}

	history := make([]PricePoint, 0)
	var point PricePoint

	for iter.Scan(&point.TokenID, &point.Timestamp, &point.Price) {
		history = append(history, point)
		point = PricePoint{} // Reset for next iteration
	}

	if err := iter.Close(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch price history"})
	}

	if len(history) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No price history found"})
	}

	return c.JSON(fiber.Map{
		"token_id": tokenID,
		"count":    len(history),
		"history":  history,
	})
}

// Get analytics from ElasticSearch
func (h *Handler) GetAnalytics(c *fiber.Ctx) error {
	var buf bytes.Buffer

	// Aggregation query
	aggQuery := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"avg_price": map[string]interface{}{
				"avg": map[string]interface{}{
					"field": "current_price",
				},
			},
			"total_market_cap": map[string]interface{}{
				"sum": map[string]interface{}{
					"field": "market_cap",
				},
			},
			"top_tokens": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "symbol",
					"size":  10,
					"order": map[string]interface{}{
						"by_market_cap": "desc",
					},
				},
				"aggs": map[string]interface{}{
					"by_market_cap": map[string]interface{}{
						"max": map[string]interface{}{
							"field": "market_cap",
						},
					},
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(aggQuery); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to encode query"})
	}

	res, err := h.ElasticSearch.Client.Search(
		h.ElasticSearch.Client.Search.WithContext(context.Background()),
		h.ElasticSearch.Client.Search.WithIndex("crypto_tokens"),
		h.ElasticSearch.Client.Search.WithBody(&buf),
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Search failed"})
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode response"})
	}

	aggs := result["aggregations"].(map[string]interface{})

	return c.JSON(fiber.Map{
		"analytics": aggs,
	})
}

// Get all tokens from ScyllaDB
func (h *Handler) GetAllTokens(c *fiber.Ctx) error {
	query := `SELECT id, symbol, name, current_price, market_cap, volume_24h, updated_at FROM tokens`

	iter := h.ScyllaDB.Session.Query(query).Iter()

	tokens := make([]models.Token, 0)
	var token models.Token

	for iter.Scan(&token.ID, &token.Symbol, &token.Name, &token.CurrentPrice,
		&token.MarketCap, &token.Volume24h, &token.UpdatedAt) {
		tokens = append(tokens, token)
		token = models.Token{} // Reset for next iteration
	}

	if err := iter.Close(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tokens"})
	}

	return c.JSON(fiber.Map{
		"tokens": tokens,
		"count":  len(tokens),
	})
}
