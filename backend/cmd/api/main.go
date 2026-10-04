package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if dbURL == "" || redisURL == "" {
		log.Fatal("ERROR: DATABASE_URL or REDIS_URL missing in .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// 1. PostgreSQL (Supabase) Test
	fmt.Println("Connecting to Supabase PostgreSQL...")
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize Postgres pool: %v", err)
	}
	defer pool.Close()

	var tripCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM trips;").Scan(&tripCount)
	if err != nil {
		log.Fatalf("Postgres ping failed: %v", err)
	}
	fmt.Printf("✅ PostgreSQL Connected successfully! Found %d scheduled trips.\n", tripCount)

	// 2. Redis (Upstash) Test
	fmt.Println("Connecting to Upstash Redis...")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Invalid Redis URL: %v", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis ping failed: %v", err)
	}
	fmt.Printf("✅ Upstash Redis Connected successfully! Response: %s\n", pong)

	fmt.Println("\n🎉 STEP 1 100% COMPLETE! Storage, Cloud DBs, and Schemas are active.")
}