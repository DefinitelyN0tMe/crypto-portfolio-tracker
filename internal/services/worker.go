package services

import (
	"context"
	"crypto-portfolio-tracker/internal/db"
	"log"
	"time"
)

type PriceWorker struct {
	ScyllaDB      *db.ScyllaDB
	ElasticSearch *db.ElasticSearch
	CoinGecko     *CoinGeckoClient
	Interval      time.Duration
}

func NewPriceWorker(scylla *db.ScyllaDB, es *db.ElasticSearch, interval time.Duration) *PriceWorker {
	return &PriceWorker{
		ScyllaDB:      scylla,
		ElasticSearch: es,
		CoinGecko:     NewCoinGeckoClient(),
		Interval:      interval,
	}
}

// Start begins the background worker
func (w *PriceWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	log.Printf("🔄 Price worker started (interval: %v)", w.Interval)

	// Initial sync on startup
	w.syncPrices()

	for {
		select {
		case <-ticker.C:
			w.syncPrices()
		case <-ctx.Done():
			log.Println("🛑 Price worker stopped")
			return
		}
	}
}

func (w *PriceWorker) syncPrices() {
	log.Println("📊 Syncing prices from CoinGecko...")

	tokens, err := w.CoinGecko.FetchTopTokens(10)
	if err != nil {
		log.Printf("❌ Failed to fetch tokens: %v", err)
		return
	}

	successCount := 0
	for _, token := range tokens {
		// Update ScyllaDB
		query := `INSERT INTO tokens (id, symbol, name, current_price, market_cap, volume_24h, updated_at) 
                  VALUES (?, ?, ?, ?, ?, ?, ?)`

		if err := w.ScyllaDB.Session.Query(query,
			token.ID, token.Symbol, token.Name, token.CurrentPrice,
			token.MarketCap, token.Volume24h, token.UpdatedAt).Exec(); err != nil {
			log.Printf("❌ Failed to update %s in ScyllaDB: %v", token.ID, err)
			continue
		}

		// Update ElasticSearch
		tokenMap := map[string]interface{}{
			"id":            token.ID,
			"symbol":        token.Symbol,
			"name":          token.Name,
			"current_price": token.CurrentPrice,
			"market_cap":    token.MarketCap,
			"volume_24h":    token.Volume24h,
			"updated_at":    token.UpdatedAt,
		}

		if err := w.ElasticSearch.IndexToken(context.Background(), tokenMap); err != nil {
			log.Printf("❌ Failed to index %s in ElasticSearch: %v", token.ID, err)
			continue
		}

		// Save to price_history
		historyQuery := `INSERT INTO price_history (token_id, timestamp, price) VALUES (?, ?, ?)`
		if err := w.ScyllaDB.Session.Query(historyQuery,
			token.ID, token.UpdatedAt, token.CurrentPrice).Exec(); err != nil {
			log.Printf("❌ Failed to save price history for %s: %v", token.ID, err)
		}

		successCount++
	}

	log.Printf("✅ Synced %d/%d tokens", successCount, len(tokens))
}
