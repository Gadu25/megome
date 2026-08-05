package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeminiGenerateTextSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-goog-api-key") == "" {
			t.Error("expected api key in X-goog-api-key header")
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"candidates":[{"content":{"parts":[{"text":"{\"tagline\":\"hi\"}"}]}}]}`)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash")
	client.baseURL = server.URL

	text, err := client.GenerateText(context.Background(), "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if text != `{"tagline":"hi"}` {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestGeminiGenerateTextQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"error":{"code":429,"message":"RESOURCE_EXHAUSTED"}}`)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash")
	client.baseURL = server.URL

	_, err := client.GenerateText(context.Background(), "prompt")
	var qe *QuotaError
	if !errors.As(err, &qe) {
		t.Fatalf("expected QuotaError, got %v", err)
	}
}

func TestGeminiGenerateTextServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `internal error`)
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash")
	client.baseURL = server.URL

	_, err := client.GenerateText(context.Background(), "prompt")
	var qe *QuotaError
	if errors.As(err, &qe) {
		t.Fatalf("expected generic error, got QuotaError: %v", err)
	}
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGeminiGenerateTextEmptyKey(t *testing.T) {
	client := NewGeminiClient("", "gemini-2.5-flash")
	_, err := client.GenerateText(context.Background(), "prompt")
	var qe *QuotaError
	if !errors.As(err, &qe) {
		t.Fatalf("expected QuotaError when key empty, got %v", err)
	}
}
