package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jaydyntfavors/Go-Ledger-AI/internal/models"
	"github.com/jaydyntfavors/Go-Ledger-Ai/internal/services"
)

type LedgerRepository struct {
	db *sql.DB
}

func NewLedgerRepository(db *sql.DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

// CreateEntry inserts a new spending transaction into ledger_entries
func (r *LedgerRepository) CreateEntry(ctx context.Context, entry *models.LedgerEntry) error {
	query := `
		INSERT INTO ledger_entries (user_id, category_id, amount_cents, merchant, description, transaction_date)
		VALUES ($1, $2, $3, $4, $5, COALESCE(NULLIF($6, '')::DATE, CURRENT_DATE))
		RETURNING id, transaction_date::TEXT, created_at
	`
	return r.db.QueryRowContext(
		ctx,
		query,
		entry.UserID,
		entry.CategoryID,
		entry.AmountCents,
		entry.Merchant,
		entry.Description,
		entry.TransactionDate,
	).Scan(&entry.ID, &entry.TransactionDate, &entry.CreatedAt)
}

// UpsertBudgetTarget allows a user to set or update a custom monthly target for a category
func (r *LedgerRepository) UpsertBudgetTarget(ctx context.Context, userID string, categoryID int, targetCents int64) error {
	query := `
		INSERT INTO budget_targets (user_id, category_id, target_cents)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, category_id)
		DO UPDATE SET target_cents = EXCLUDED.target_cents
	`
	_, err := r.db.ExecContext(ctx, query, userID, categoryID, targetCents)
	return err
}

// GetUserTakeHome fetches the user's monthly income in cents for the 50/30/20 calculation
func (r *LedgerRepository) GetUserTakeHome(ctx context.Context, userID string) (int64, error) {
	var takeHomeCents int64
	err := r.db.QueryRowContext(ctx, `SELECT monthly_take_home_cents FROM users WHERE id = $1`, userID).Scan(&takeHomeCents)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch user take-home pay: %w", err)
	}
	return takeHomeCents, nil
}

// GetMonthlyCategoryInputs aggregates monthly spending per category alongside any custom targets
func (r *LedgerRepository) GetMonthlyCategoryInputs(ctx context.Context, userID string, year int, month int) ([]services.CategoryInput, error) {
	query := `
		SELECT 
			c.name,
			c.bucket,
			COALESCE(SUM(le.amount_cents), 0) AS actual_spent_cents,
			bt.target_cents
		FROM categories c
		LEFT JOIN ledger_entries le 
			ON le.category_id = c.id 
			AND le.user_id = $1
			AND EXTRACT(YEAR FROM le.transaction_date) = $2
			AND EXTRACT(MONTH FROM le.transaction_date) = $3
		LEFT JOIN budget_targets bt 
			ON bt.category_id = c.id 
			AND bt.user_id = $1
		GROUP BY c.id, c.name, c.bucket, bt.target_cents
		ORDER BY c.id;
	`

	rows, err := r.db.QueryContext(ctx, query, userID, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inputs []services.CategoryInput
	for rows.Next() {
		var in services.CategoryInput
		var bucketStr string
		var customTarget sql.NullInt64

		if err := rows.Scan(&in.CategoryName, &bucketStr, &in.ActualSpentCents, &customTarget); err != nil {
			return nil, err
		}

		in.Bucket = models.BucketType(bucketStr)
		if customTarget.Valid {
			val := customTarget.Int64
			in.CustomTargetCents = &val
		}

		inputs = append(inputs, in)
	}

	return inputs, rows.Err()
}
