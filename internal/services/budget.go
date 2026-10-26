package services

import "github.com/jaydyntfavors-tech/Go-Ledger-AI/internal/models"

// CategoryInput holds raw data from the database before we run our budget math
type CategoryInput struct {
	CategoryName      string
	Bucket            models.BucketType
	ActualSpentCents  int64
	CustomTargetCents *int64 // Pointer so we know if it's nil (not set by user)
}

// EvaluateMonthlyBudget runs either Custom Target checks or the 50/30/20 fallback rule
func EvaluateMonthlyBudget(userID string, year int, month int, takeHomeCents int64, inputs []CategoryInput) models.MonthlyReport {
	var totalSpent int64
	var hasCustomTargets bool

	for _, in := range inputs {
		totalSpent += in.ActualSpentCents
		if in.CustomTargetCents != nil {
			hasCustomTargets = true
		}
	}

	var breakdown []models.CategoryStatus

	if hasCustomTargets {
		// Mode 1: User configured custom targets per category
		for _, in := range inputs {
			var target int64
			rule := "50_30_20_FALLBACK"
			if in.CustomTargetCents != nil {
				target = *in.CustomTargetCents
				rule = "CUSTOM_TARGET"
			}

			delta := target - in.ActualSpentCents
			breakdownItem := models.CategoryStatus{
				CategoryName: in.CategoryName,
				Bucket:       in.Bucket,
				TargetCents:  target,
				ActualCents:  in.ActualSpentCents,
				DeltaCents:   delta,
				Status:       determineStatus(delta),
				RuleApplied:  rule,
			}
			breakdown = append(breakdown, breakdownItem)
		}
	} else {
		// Mode 2: Pure 50/30/20 Fallback grouped by NEEDS (50%), WANTS (30%), SAVINGS (20%)
		bucketSpent := map[models.BucketType]int64{
			models.BucketNeeds:   0,
			models.BucketWants:   0,
			models.BucketSavings: 0,
		}

		for _, in := range inputs {
			bucketSpent[in.Bucket] += in.ActualSpentCents
		}

		bucketTargets := map[models.BucketType]int64{
			models.BucketNeeds:   (takeHomeCents * 50) / 100,
			models.BucketWants:   (takeHomeCents * 30) / 100,
			models.BucketSavings: (takeHomeCents * 20) / 100,
		}

		bucketsOrder := []models.BucketType{models.BucketNeeds, models.BucketWants, models.BucketSavings}
		for _, b := range bucketsOrder {
			target := bucketTargets[b]
			actual := bucketSpent[b]
			delta := target - actual

			breakdown = append(breakdown, models.CategoryStatus{
				CategoryName: string(b) + " Bucket",
				Bucket:       b,
				TargetCents:  target,
				ActualCents:  actual,
				DeltaCents:   delta,
				Status:       determineStatus(delta),
				RuleApplied:  "50_30_20_FALLBACK",
			})
		}
	}

	return models.MonthlyReport{
		UserID:               userID,
		Year:                 year,
		Month:                month,
		MonthlyTakeHomeCents: takeHomeCents,
		TotalSpentCents:      totalSpent,
		TotalRemainingCents:  takeHomeCents - totalSpent,
		Breakdown:            breakdown,
	}
}

func determineStatus(deltaCents int64) string {
	if deltaCents < 0 {
		return "OVER_TARGET"
	} else if deltaCents == 0 {
		return "ON_TARGET"
	}
	return "UNDER_TARGET"
}
