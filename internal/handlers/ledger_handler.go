package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jaydyntfavors-tech/Go-Ledger-AI/internal/models"
	"github.com/jaydyntfavors-tech/Go-Ledger-AI/internal/repository"
	"github.com/jaydyntfavors-tech/Go-Ledger-AI/internal/services"
)

type LedgerHandler struct {
	repo    *repository.LedgerRepository
	advisor *services.AIAdvisor
}

func NewLedgerHandler(repo *repository.LedgerRepository, advisor *services.AIAdvisor) *LedgerHandler {
	return &LedgerHandler{
		repo:    repo,
		advisor: advisor,
	}
}

// CreateTransaction handles POST /api/v1/transactions
func (h *LedgerHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var entry models.LedgerEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, `{"error": "Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	if entry.UserID == "" || entry.CategoryID == 0 || entry.AmountCents <= 0 || entry.Merchant == "" {
		http.Error(w, `{"error": "user_id, category_id, positive amount_cents, and merchant are required"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.CreateEntry(r.Context(), &entry); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// SetBudgetTarget handles PUT /api/v1/budgets/targets
func (h *LedgerHandler) SetBudgetTarget(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      string `json:"user_id"`
		CategoryID  int    `json:"category_id"`
		TargetCents int64  `json:"target_cents"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpsertBudgetTarget(r.Context(), req.UserID, req.CategoryID, req.TargetCents); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "target_updated"})
}

// GetMonthlyReport handles GET /api/v1/reports/{year}/{month}?user_id=...&include_ai=true
func (h *LedgerHandler) GetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, `{"error": "user_id query parameter is required"}`, http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(r.PathValue("year"))
	if err != nil {
		http.Error(w, `{"error": "invalid year"}`, http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(r.PathValue("month"))
	if err != nil || month < 1 || month > 12 {
		http.Error(w, `{"error": "invalid month (1-12)"}`, http.StatusBadRequest)
		return
	}

	takeHomeCents, err := h.repo.GetUserTakeHome(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	inputs, err := h.repo.GetMonthlyCategoryInputs(r.Context(), userID, year, month)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	report := services.EvaluateMonthlyBudget(userID, year, month, takeHomeCents, inputs)

	// If ?include_ai=true is passed, attach Gemini's financial analysis
	if r.URL.Query().Get("include_ai") == "true" {
		insights, err := h.advisor.GenerateMonthlyInsights(r.Context(), report)
		if err != nil {
			report.AIInsights = "Could not generate AI insights: " + err.Error()
		} else {
			report.AIInsights = insights
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}
