package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type QuotaError struct{ msg string }

func (e *QuotaError) Error() string { return e.msg }

type GeminiClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	return &GeminiClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *GeminiClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", &QuotaError{msg: "ai is not configured"}
	}

	body := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]string{{"text": prompt}}},
		},
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
			"maxOutputTokens":  500,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read gemini response: %w", err)
	}

	if res.StatusCode == http.StatusTooManyRequests {
		return "", &QuotaError{msg: fmt.Sprintf("gemini quota exceeded: %s", respBody)}
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API error: status %d, body %s", res.StatusCode, respBody)
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("parse gemini response: %w", err)
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty gemini response")
	}

	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
