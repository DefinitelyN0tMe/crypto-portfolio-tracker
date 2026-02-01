package models

import "time"

// Token represents a cryptocurrency token
type Token struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	CurrentPrice float64   `json:"current_price"`
	MarketCap    float64   `json:"market_cap"`
	Volume24h    float64   `json:"volume_24h"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PriceHistory stores historical price data
type PriceHistory struct {
	TokenID   string    `json:"token_id"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// Portfolio represents user's crypto holdings
type Portfolio struct {
	UserID   string    `json:"user_id"`
	TokenID  string    `json:"token_id"`
	Amount   float64   `json:"amount"`
	BuyPrice float64   `json:"buy_price"`
	BuyDate  time.Time `json:"buy_date"`
}
