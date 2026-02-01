package services

import (
	"crypto-portfolio-tracker/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CoinGeckoClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewCoinGeckoClient() *CoinGeckoClient {
	return &CoinGeckoClient{
		BaseURL: "https://api.coingecko.com/api/v3",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CoinGecko API response structure
type CoinGeckoToken struct {
	ID             string  `json:"id"`
	Symbol         string  `json:"symbol"`
	Name           string  `json:"name"`
	CurrentPrice   float64 `json:"current_price"`
	MarketCap      float64 `json:"market_cap"`
	TotalVolume    float64 `json:"total_volume"`
	PriceChange24h float64 `json:"price_change_24h"`
}

// Fetch top tokens by market cap
func (c *CoinGeckoClient) FetchTopTokens(limit int) ([]models.Token, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1&sparkline=false",
		c.BaseURL, limit)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s (status %d)", string(body), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cgTokens []CoinGeckoToken
	if err := json.Unmarshal(body, &cgTokens); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to our Token model
	tokens := make([]models.Token, 0, len(cgTokens))
	for _, cg := range cgTokens {
		tokens = append(tokens, models.Token{
			ID:           cg.ID,
			Symbol:       cg.Symbol,
			Name:         cg.Name,
			CurrentPrice: cg.CurrentPrice,
			MarketCap:    cg.MarketCap,
			Volume24h:    cg.TotalVolume,
			UpdatedAt:    time.Now(),
		})
	}

	return tokens, nil
}

// Fetch single token by ID
func (c *CoinGeckoClient) FetchToken(tokenID string) (*models.Token, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&ids=%s", c.BaseURL, tokenID)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cgTokens []CoinGeckoToken
	if err := json.Unmarshal(body, &cgTokens); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(cgTokens) == 0 {
		return nil, fmt.Errorf("token not found: %s", tokenID)
	}

	cg := cgTokens[0]
	return &models.Token{
		ID:           cg.ID,
		Symbol:       cg.Symbol,
		Name:         cg.Name,
		CurrentPrice: cg.CurrentPrice,
		MarketCap:    cg.MarketCap,
		Volume24h:    cg.TotalVolume,
		UpdatedAt:    time.Now(),
	}, nil
}
