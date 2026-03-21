package classification

import (
	"context"
	"strings"
)

const (
	TAG_INVALID    = "TAG_INVALID"
	TAG_NOT_FOUND  = "TAG_NOT_FOUND"
	TAG_DUPLICATED = "TAG_DUPLICATED"
)

type TagService struct {
	repo Repository
}

func NewTagService(repo Repository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) CreateTag(_ context.Context, userID string, input TagCreateInput) (Tag, error) {
	normalizedUserID := strings.TrimSpace(userID)
	name := strings.TrimSpace(input.Name)
	if normalizedUserID == "" || name == "" {
		return Tag{}, &contractError{code: TAG_INVALID}
	}

	if _, found, err := s.repo.GetTagByNameForUser(normalizedUserID, name); err != nil {
		return Tag{}, err
	} else if found {
		return Tag{}, &contractError{code: TAG_DUPLICATED}
	}

	input.Name = name
	return s.repo.CreateTag(normalizedUserID, input)
}

func (s *TagService) ListTags(_ context.Context, userID string) ([]Tag, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, &contractError{code: TAG_INVALID}
	}
	return s.repo.ListTagsByUser(normalizedUserID)
}

func (s *TagService) UpdateTag(_ context.Context, userID string, tagID string, input TagUpdateInput) (Tag, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedTagID := strings.TrimSpace(tagID)
	name := strings.TrimSpace(input.Name)
	if normalizedUserID == "" || normalizedTagID == "" || name == "" {
		return Tag{}, &contractError{code: TAG_INVALID}
	}

	tag, found, err := s.repo.GetTagByIDForUser(normalizedUserID, normalizedTagID)
	if err != nil {
		return Tag{}, err
	}
	if !found {
		return Tag{}, &contractError{code: TAG_NOT_FOUND}
	}

	if existing, dup, dupErr := s.repo.GetTagByNameForUser(normalizedUserID, name); dupErr != nil {
		return Tag{}, dupErr
	} else if dup && existing.ID != tag.ID {
		return Tag{}, &contractError{code: TAG_DUPLICATED}
	}

	tag.Name = name
	updated, saved, saveErr := s.repo.SaveTagByIDForUser(normalizedUserID, normalizedTagID, tag)
	if saveErr != nil {
		return Tag{}, saveErr
	}
	if !saved {
		return Tag{}, &contractError{code: TAG_NOT_FOUND}
	}
	return updated, nil
}

func (s *TagService) DeleteTag(_ context.Context, userID string, tagID string) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedTagID := strings.TrimSpace(tagID)
	if normalizedUserID == "" || normalizedTagID == "" {
		return &contractError{code: TAG_INVALID}
	}

	deleted, err := s.repo.DeleteTagByIDForUser(normalizedUserID, normalizedTagID)
	if err != nil {
		return err
	}
	if !deleted {
		return &contractError{code: TAG_NOT_FOUND}
	}
	return nil
}

func (s *TagService) AttachTagToTransaction(_ context.Context, userID string, tagID string, transactionID string) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedTagID := strings.TrimSpace(tagID)
	normalizedTransactionID := strings.TrimSpace(transactionID)
	if normalizedUserID == "" || normalizedTagID == "" || normalizedTransactionID == "" {
		return &contractError{code: TAG_INVALID}
	}

	if _, found, err := s.repo.GetTagByIDForUser(normalizedUserID, normalizedTagID); err != nil {
		return err
	} else if !found {
		return &contractError{code: TAG_NOT_FOUND}
	}

	return s.repo.AttachTagToTransaction(normalizedUserID, normalizedTagID, normalizedTransactionID)
}

func (s *TagService) ValidateTagIDs(_ context.Context, userID string, tagIDs []string) error {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return &contractError{code: TAG_INVALID}
	}
	for _, tagID := range tagIDs {
		normalizedTagID := strings.TrimSpace(tagID)
		if normalizedTagID == "" {
			return &contractError{code: TAG_INVALID}
		}
		if _, found, err := s.repo.GetTagByIDForUser(normalizedUserID, normalizedTagID); err != nil {
			return err
		} else if !found {
			return &contractError{code: TAG_NOT_FOUND}
		}
	}
	return nil
}

func (s *TagService) ReplaceTransactionTags(ctx context.Context, userID string, transactionID string, tagIDs []string) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedTransactionID := strings.TrimSpace(transactionID)
	if normalizedUserID == "" || normalizedTransactionID == "" {
		return &contractError{code: TAG_INVALID}
	}
	cleaned := make([]string, 0, len(tagIDs))
	seen := map[string]bool{}
	for _, tagID := range tagIDs {
		normalizedTagID := strings.TrimSpace(tagID)
		if normalizedTagID == "" {
			return &contractError{code: TAG_INVALID}
		}
		if seen[normalizedTagID] {
			continue
		}
		seen[normalizedTagID] = true
		cleaned = append(cleaned, normalizedTagID)
	}
	if err := s.ValidateTagIDs(ctx, normalizedUserID, cleaned); err != nil {
		return err
	}
	return s.repo.ReplaceTransactionTags(normalizedUserID, normalizedTransactionID, cleaned)
}

func (s *TagService) RemoveTransactionTags(_ context.Context, userID string, transactionID string) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedTransactionID := strings.TrimSpace(transactionID)
	if normalizedUserID == "" || normalizedTransactionID == "" {
		return &contractError{code: TAG_INVALID}
	}
	return s.repo.RemoveTransactionTags(normalizedUserID, normalizedTransactionID)
}

func (s *TagService) ListTransactionIDsByTag(_ context.Context, userID string, tagID string) ([]string, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedTagID := strings.TrimSpace(tagID)
	if normalizedUserID == "" || normalizedTagID == "" {
		return nil, &contractError{code: TAG_INVALID}
	}

	if _, found, err := s.repo.GetTagByIDForUser(normalizedUserID, normalizedTagID); err != nil {
		return nil, err
	} else if !found {
		return nil, &contractError{code: TAG_NOT_FOUND}
	}

	return s.repo.ListTransactionIDsByTag(normalizedUserID, normalizedTagID)
}
