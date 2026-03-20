package classification

import (
	"context"
	"strings"
	"sync"
)

type TemplateRepository interface {
	CopyDefaultTemplateToUser(ctx context.Context, userID string) (bool, error)
}

type TemplateService struct {
	repo TemplateRepository
}

func NewTemplateService(repo TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

func (s *TemplateService) EnsureUserDefaults(ctx context.Context, userID string) error {
	if s == nil || s.repo == nil {
		return nil
	}
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return nil
	}
	_, err := s.repo.CopyDefaultTemplateToUser(ctx, trimmedUserID)
	return err
}

type InMemoryTemplateRepository struct {
	mu      sync.Mutex
	copied  map[string]bool
	copyCnt map[string]int
}

func NewInMemoryTemplateRepository() *InMemoryTemplateRepository {
	return &InMemoryTemplateRepository{
		copied:  map[string]bool{},
		copyCnt: map[string]int{},
	}
}

func (r *InMemoryTemplateRepository) CopyDefaultTemplateToUser(_ context.Context, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.copied[userID] {
		return false, nil
	}
	r.copied[userID] = true
	r.copyCnt[userID]++
	return true, nil
}

func (r *InMemoryTemplateRepository) CopyCount(userID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.copyCnt[strings.TrimSpace(userID)]
}
