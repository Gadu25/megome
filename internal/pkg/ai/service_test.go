package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"megome/internal/domain/assist"
)

type fakeProvider struct {
	text string
	err  error
}

func (f *fakeProvider) GenerateText(ctx context.Context, prompt string) (string, error) {
	return f.text, f.err
}

func TestAssistSuccess(t *testing.T) {
	status := NewStatusTracker(true, time.Minute)
	svc := NewService(&fakeProvider{text: `{"description":"Built a thing"}`}, status)

	fields, err := svc.Assist(context.Background(), string(assist.TaskGenerateEducation), map[string]string{"degree": "BS CS"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if fields["description"] != "Built a thing" {
		t.Errorf("unexpected fields: %v", fields)
	}
	if available, _ := svc.Status(); !available {
		t.Error("expected status to remain available after success")
	}
}

func TestAssistUnknownTask(t *testing.T) {
	status := NewStatusTracker(true, time.Minute)
	svc := NewService(&fakeProvider{text: `{}`}, status)

	_, err := svc.Assist(context.Background(), "bogus_task", nil, "")
	if !errors.Is(err, ErrUnknownTask) {
		t.Fatalf("expected ErrUnknownTask, got %v", err)
	}
}

func TestAssistQuotaFlipsStatus(t *testing.T) {
	status := NewStatusTracker(true, time.Minute)
	svc := NewService(&fakeProvider{err: &QuotaError{msg: "quota"}}, status)

	_, err := svc.Assist(context.Background(), string(assist.TaskGenerateBio), nil, "")
	var ua *UnavailableError
	if !errors.As(err, &ua) {
		t.Fatalf("expected UnavailableError, got %v", err)
	}
	if available, _ := svc.Status(); available {
		t.Error("expected status unavailable after quota error")
	}
}

func TestAssistBlockedWhenUnavailable(t *testing.T) {
	status := NewStatusTracker(true, time.Minute)
	status.MarkUnavailable()
	provider := &fakeProvider{text: `{"tagline":"x"}`}
	svc := NewService(provider, status)

	_, err := svc.Assist(context.Background(), string(assist.TaskGenerateBio), nil, "")
	var ua *UnavailableError
	if !errors.As(err, &ua) {
		t.Fatalf("expected UnavailableError when already unavailable, got %v", err)
	}
}

func TestAssistGenerationError(t *testing.T) {
	status := NewStatusTracker(true, time.Minute)
	svc := NewService(&fakeProvider{err: errors.New("network down")}, status)

	_, err := svc.Assist(context.Background(), string(assist.TaskGenerateBio), nil, "")
	if !errors.Is(err, ErrGeneration) {
		t.Fatalf("expected ErrGeneration, got %v", err)
	}
	if available, _ := svc.Status(); !available {
		t.Error("non-quota errors must not flip status")
	}
}
