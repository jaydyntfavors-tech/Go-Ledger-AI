package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jaydyntfavors/Go-Ledger-AI/internal/models"
)

type AIAdvisor struct {
	apiKey     string
	httpClient *http.Client
}

func NewAIAdvisor(apiKey string) *AIAdvisor {
	return &AIAdvisor{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (a *AIAdvisor) GenerateMonthlyInsights(ctx context.Context, report models.MonthlyReport) (string, error) {
	if a.apiKey == "" {
		return "AI insights unavailable: GEMINI_API_KEY is not configured.", nil
	}

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`You are a sharp, practical financial advisor integrated into a Go ledger API.
Analyze the following monthly budget report (all monetary values are in cents, so 300000 = $3,000.00).
1. State clearly where the user went OVER_TARGET or stayed UNDER_TARGET.
2. Note whether they are being evaluated on custom category targets or the 50/30/20 rule fallback.
3. Give 2-3 specific, actionable recommendations for next month using exact dollar amounts converted from the cents provided. Keep it concise.

Monthly Report JSON:
%s`, string(reportJSON))

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", a.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rawErr, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(rawErr))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "No insights generated.", nil
}
