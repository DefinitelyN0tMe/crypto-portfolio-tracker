package db

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

type ScyllaDB struct {
	Session *gocql.Session
}

func NewScyllaDB(hosts []string) (*ScyllaDB, error) {
	// First connection without keyspace to create it
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ScyllaDB: %w", err)
	}

	log.Println("✅ Connected to ScyllaDB")
	return &ScyllaDB{Session: session}, nil
}

func (db *ScyllaDB) InitSchema() error {
	// Create keyspace
	keyspaceQuery := `
        CREATE KEYSPACE IF NOT EXISTS crypto_tracker 
        WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
    `
	if err := db.Session.Query(keyspaceQuery).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace: %w", err)
	}
	log.Println("✅ Created keyspace crypto_tracker")

	// Close initial session
	db.Session.Close()

	// Reconnect with keyspace
	cluster := gocql.NewCluster("localhost:9042")
	cluster.Keyspace = "crypto_tracker"
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to reconnect with keyspace: %w", err)
	}
	db.Session = session

	// Create tokens table
	tokensTable := `
        CREATE TABLE IF NOT EXISTS tokens (
            id text PRIMARY KEY,
            symbol text,
            name text,
            current_price double,
            market_cap double,
            volume_24h double,
            updated_at timestamp
        )
    `
	if err := db.Session.Query(tokensTable).Exec(); err != nil {
		return fmt.Errorf("failed to create tokens table: %w", err)
	}

	// Create price_history table
	priceHistoryTable := `
        CREATE TABLE IF NOT EXISTS price_history (
            token_id text,
            timestamp timestamp,
            price double,
            PRIMARY KEY (token_id, timestamp)
        ) WITH CLUSTERING ORDER BY (timestamp DESC)
    `
	if err := db.Session.Query(priceHistoryTable).Exec(); err != nil {
		return fmt.Errorf("failed to create price_history table: %w", err)
	}

	log.Println("✅ ScyllaDB schema initialized")
	return nil
}

func (db *ScyllaDB) Close() {
	if db.Session != nil {
		db.Session.Close()
	}
}
