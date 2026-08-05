package ai

import (
	"context"
	"errors"
	"log"
)

var ErrUnknownTask = errors.New("unknown ai task")
var ErrGeneration = errors.New("ai generation failed")

type UnavailableError struct {
	RemainingSeconds int64
}

func (e *UnavailableError) Error() string {
	if e.RemainingSeconds < 0 {
		return "ai assist is disabled"
	}
	return "ai assist is unavailable due to quota"
}

type Service struct {
	provider Provider
	status   *StatusTracker
}

func NewService(provider Provider, status *StatusTracker) *Service {
	return &Service{provider: provider, status: status}
}

func (s *Service) Assist(ctx context.Context, task string, context map[string]string, extra string) (map[string]string, error) {
	available, remaining := s.status.Status()
	if !available {
		return nil, &UnavailableError{RemainingSeconds: remaining}
	}
	if !KnownTask(task) {
		return nil, ErrUnknownTask
	}

	prompt := BuildPrompt(task, context, extra)
	text, err := s.provider.GenerateText(ctx, prompt)
	if err != nil {
		var qe *QuotaError
		if errors.As(err, &qe) {
			log.Printf("gemini quota error: %s", qe.msg)
			s.status.MarkUnavailable(qe.RetryAfter())
			_, remaining := s.status.Status()
			return nil, &UnavailableError{RemainingSeconds: remaining}
		}
		log.Printf("ai generation failed: %v", err)
		return nil, ErrGeneration
	}

	fields, err := ParseFields(text)
	if err != nil {
		log.Printf("ai parse fields failed (raw=%q): %v", text, err)
		return nil, ErrGeneration
	}
	return fields, nil
}

func (s *Service) Status() (bool, int64) {
	return s.status.Status()
}
