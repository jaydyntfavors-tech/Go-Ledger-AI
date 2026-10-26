package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jaydyntfavors-tech/Go-Ledger-AI/internal/config"
	"github.com/jaydyntfavors-tech/Go-Ledger-AI/internal/handlers"
	"github.com/jaydyntfavors-tech/Go-Ledger-AI/internal/repository"
	"github.com/jaydyntfavors-tech/Go-Ledger-AI/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	if cfg.DatabaseURL != "" {
		if err := db.Ping(); err != nil {
			log.Printf("WARNING: Database ping failed: %v", err)
		} else {
			log.Println("Connected to Neon PostgreSQL successfully!")
		}
	}

	repo := repository.NewLedgerRepository(db)
	advisor := services.NewAIAdvisor(cfg.GeminiAPIKey)
	ledgerHandler := handlers.NewLedgerHandler(repo, advisor)

	mux := http.NewServeMux()

	// 1. Health Check Endpoint
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "Go-Ledger-AI",
		})
	})

	// 2. Core Ledger & Budget Endpoints
	mux.HandleFunc("POST /api/v1/transactions", ledgerHandler.CreateTransaction)
	mux.HandleFunc("PUT /api/v1/budgets/targets", ledgerHandler.SetBudgetTarget)
	mux.HandleFunc("GET /api/v1/reports/{year}/{month}", ledgerHandler.GetMonthlyReport)

	log.Printf("Starting Go-Ledger-AI server on port %s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
