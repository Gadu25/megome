package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"megome/internal/domain/user"
	"megome/internal/pkg/ai"

	"github.com/gorilla/mux"
)

type fakeAIProvider struct {
	text string
	err  error
}

func (f *fakeAIProvider) GenerateText(ctx context.Context, prompt string) (string, error) {
	return f.text, f.err
}

func newAssistTestHandler(provider ai.Provider, status *ai.StatusTracker) *AssistHandler {
	return NewAssistHandler(ai.NewService(provider, status), user.NewRepository(nil))
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v; body=%s", err, rr.Body.String())
	}
	return body
}

func TestAssistHandlerAssistSuccess(t *testing.T) {
	status := ai.NewStatusTracker(true, time.Minute)
	h := newAssistTestHandler(&fakeAIProvider{text: `{"description":"Built a thing"}`}, status)

	req := httptest.NewRequest(http.MethodPost, "/ai/assist", strings.NewReader(`{"task":"generate_bio","context":{"name":"x"},"extra":""}`))
	rr := httptest.NewRecorder()

	h.handleAssist(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", rr.Code, rr.Body.String())
	}
	body := decodeBody(t, rr)
	data := body["data"].(map[string]any)
	if data["task"] != "generate_bio" {
		t.Errorf("expected task echoed in response, got %v", data["task"])
	}
	fields := data["fields"].(map[string]any)
	if fields["description"] != "Built a thing" {
		t.Errorf("expected generated fields in response, got %v", fields)
	}
}

func TestAssistHandlerAssistUnavailable(t *testing.T) {
	status := ai.NewStatusTracker(true, time.Hour)
	status.MarkUnavailable()
	h := newAssistTestHandler(&fakeAIProvider{text: `{"description":"x"}`}, status)

	req := httptest.NewRequest(http.MethodPost, "/ai/assist", strings.NewReader(`{"task":"generate_bio"}`))
	rr := httptest.NewRecorder()

	h.handleAssist(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d; body=%s", rr.Code, rr.Body.String())
	}
	body := decodeBody(t, rr)
	if body["message"] != "ai_unavailable" {
		t.Errorf("expected ai_unavailable message, got %v", body["message"])
	}
	data := body["data"].(map[string]any)
	remaining, ok := data["cooldownRemainingSeconds"].(float64)
	if !ok || remaining <= 0 {
		t.Errorf("expected positive cooldownRemainingSeconds, got %v", data["cooldownRemainingSeconds"])
	}
}

func TestAssistHandlerAssistUnknownTask(t *testing.T) {
	status := ai.NewStatusTracker(true, time.Minute)
	h := newAssistTestHandler(&fakeAIProvider{text: `{"description":"x"}`}, status)

	req := httptest.NewRequest(http.MethodPost, "/ai/assist", strings.NewReader(`{"task":"bogus"}`))
	rr := httptest.NewRecorder()

	h.handleAssist(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d; body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "unknown ai task") {
		t.Errorf("expected unknown task error in body, got %s", rr.Body.String())
	}
}

func TestAssistHandlerAssistGenerationError(t *testing.T) {
	status := ai.NewStatusTracker(true, time.Minute)
	h := newAssistTestHandler(&fakeAIProvider{err: errors.New("network down")}, status)

	req := httptest.NewRequest(http.MethodPost, "/ai/assist", strings.NewReader(`{"task":"generate_bio"}`))
	rr := httptest.NewRecorder()

	h.handleAssist(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d; body=%s", rr.Code, rr.Body.String())
	}
}

func TestAssistHandlerStatus(t *testing.T) {
	status := ai.NewStatusTracker(true, time.Minute)
	h := newAssistTestHandler(&fakeAIProvider{text: `{}`}, status)

	req := httptest.NewRequest(http.MethodGet, "/ai/status", nil)
	rr := httptest.NewRecorder()

	h.handleStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", rr.Code, rr.Body.String())
	}
	body := decodeBody(t, rr)
	data := body["data"].(map[string]any)
	if data["available"] != true {
		t.Errorf("expected available=true, got %v", data["available"])
	}
	if data["cooldownRemainingSeconds"] != float64(0) {
		t.Errorf("expected cooldownRemainingSeconds=0, got %v", data["cooldownRemainingSeconds"])
	}
}

func TestAssistHandlerRegisterRoutes(t *testing.T) {
	h := newAssistTestHandler(&fakeAIProvider{text: `{}`}, ai.NewStatusTracker(true, time.Minute))

	router := mux.NewRouter()
	h.RegisterRoutes(router)

	var match mux.RouteMatch

	postAssist := httptest.NewRequest(http.MethodPost, "/ai/assist", strings.NewReader(`{"task":"generate_bio"}`))
	if !router.Match(postAssist, &match) {
		t.Error("expected POST /ai/assist to be registered")
	}

	getStatus := httptest.NewRequest(http.MethodGet, "/ai/status", nil)
	if !router.Match(getStatus, &match) {
		t.Error("expected GET /ai/status to be registered")
	}

	wrongMethod := httptest.NewRequest(http.MethodGet, "/ai/assist", nil)
	if router.Match(wrongMethod, &match) {
		t.Error("GET /ai/assist should not match")
	}
}
