package models

import "time"

// BucketType reperesents the 50/30/20 rule classifications
type BucketType string

const(
  BucketNeeds  BucketType = "NEEDS"  // 50% default
  BucketWants  BucketType = "WANTS"  // 30% default
  BucketSavings  BucketType = "SAVINGS"  // 20% default
  )

// LedgerEntry represents a single financial transaction
type LedgerEntry struct{
  ID              string    'json:"id"'
  UserID          string    'json:"user_id"'
  CategoryID      int       'json:"category_id"'
  CatergoryName   string    'json:"category_name,omitempty"'
  AmountCents     int64     'json:"amount_cents"'
  Merchant        string    'json:"merchant"'
  Description     string    'json:"description,omitempty"'
  TransactionDate string    'json:"transaction_date"' // YYYY-MM-DD
  CreatedAT       time.Time 'json:"created_at"'
}

// CategoryStatus shows how user performed in a certain catergory or bucket
type CategoryStatus struct {
  CategoryName      string      'json:"category_name"'
  Bucket            BucketType  'json:"bucket"'
  TargetCents       int64       'json:"target_cents"'
  ActualCents       int64       'json:"actual_cents"'
  DeltaCents        int64       'json:"delta_cents"' // Target - Actual (Negative means OVER budget)
  Status            string      'json:"status"'      // "UNDER_TARGET", "ON_TARGET", or "OVER_TARGET"
  RuleApplied       string      'json:"rule_applied"'// "CUSTOM_TARGET" or "50_30_20_FALLBACK"
}

// MonthlyReports is the full payload returned by the monthly endpoint
type MonthlyReport struct{
  UserID                      string            'json:"user_id"'
  Year                        int               'json:"year"'
  Month                       int               'json:"month"'
  MonthlyTakeHomeCents        int64             'json:"monthly_take_home_cents"'
  TotalSpentCents             int64             'json:"total_spent_cents"'
  TotalRemainingCents         int64             'json:"total_remaining_cents"'
  Breakdown                   []CategoryStatus  'json:"ai_insights,omitempty"'
}
