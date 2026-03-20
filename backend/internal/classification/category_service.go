package classification

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	CAT_INVALID         = "CAT_INVALID"
	CAT_NOT_FOUND       = "CAT_NOT_FOUND"
	CAT_INVALID_PARENT  = "CAT_INVALID_PARENT"
	CAT_IN_USE_ARCHIVED = "CAT_IN_USE_ARCHIVED"
	CAT_ARCHIVED        = "CAT_ARCHIVED"
)

type contractError struct {
	code string
	err  error
}

func (e *contractError) Error() string {
	if e.err == nil {
		return e.code
	}
	return e.code + ": " + e.err.Error()
}

func (e *contractError) Unwrap() error {
	return e.err
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	var coded *contractError
	if errors.As(err, &coded) {
		return coded.code
	}
	return ""
}

type CategoryService struct {
	repo Repository
}

type DeleteCategoryResult struct {
	Deleted  bool     `json:"deleted"`
	Archived bool     `json:"archived"`
	Category Category `json:"category"`
}

func NewCategoryService(repo Repository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(_ context.Context, userID string, input CategoryCreateInput) (Category, error) {
	normalizedUserID := strings.TrimSpace(userID)
	name := strings.TrimSpace(input.Name)
	if normalizedUserID == "" || name == "" {
		return Category{}, &contractError{code: CAT_INVALID}
	}

	input.Name = name
	input.ParentID = normalizeOptionalID(input.ParentID)
	if err := s.validateParent(normalizedUserID, "", input.ParentID); err != nil {
		return Category{}, err
	}

	return s.repo.CreateCategory(normalizedUserID, input)
}

func (s *CategoryService) ListCategories(_ context.Context, userID string) ([]Category, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, &contractError{code: CAT_INVALID}
	}
	return s.repo.ListCategoriesByUser(normalizedUserID)
}

func (s *CategoryService) UpdateCategory(_ context.Context, userID string, categoryID string, input CategoryUpdateInput) (Category, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCategoryID := strings.TrimSpace(categoryID)
	if normalizedUserID == "" || normalizedCategoryID == "" {
		return Category{}, &contractError{code: CAT_INVALID}
	}

	category, found, err := s.repo.GetCategoryByIDForUser(normalizedUserID, normalizedCategoryID)
	if err != nil {
		return Category{}, err
	}
	if !found {
		return Category{}, &contractError{code: CAT_NOT_FOUND}
	}

	if input.Name != nil {
		trimmedName := strings.TrimSpace(*input.Name)
		if trimmedName == "" {
			return Category{}, &contractError{code: CAT_INVALID}
		}
		category.Name = trimmedName
	}

	newParentID := category.ParentID
	if input.ClearParent {
		newParentID = nil
	}
	if input.ParentID != nil {
		newParentID = normalizeOptionalID(input.ParentID)
	}

	if err := s.validateParent(normalizedUserID, normalizedCategoryID, newParentID); err != nil {
		return Category{}, err
	}
	category.ParentID = newParentID

	if input.Archive != nil {
		if *input.Archive {
			now := time.Now().UTC()
			category.ArchivedAt = &now
		} else {
			category.ArchivedAt = nil
		}
	}

	updated, saved, saveErr := s.repo.SaveCategoryByIDForUser(normalizedUserID, normalizedCategoryID, category)
	if saveErr != nil {
		return Category{}, saveErr
	}
	if !saved {
		return Category{}, &contractError{code: CAT_NOT_FOUND}
	}
	return updated, nil
}

func (s *CategoryService) DeleteCategory(_ context.Context, userID string, categoryID string) (DeleteCategoryResult, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCategoryID := strings.TrimSpace(categoryID)
	if normalizedUserID == "" || normalizedCategoryID == "" {
		return DeleteCategoryResult{}, &contractError{code: CAT_INVALID}
	}

	category, found, err := s.repo.GetCategoryByIDForUser(normalizedUserID, normalizedCategoryID)
	if err != nil {
		return DeleteCategoryResult{}, err
	}
	if !found {
		return DeleteCategoryResult{}, &contractError{code: CAT_NOT_FOUND}
	}

	hasChildren, childErr := s.repo.CategoryHasChildren(normalizedUserID, normalizedCategoryID)
	if childErr != nil {
		return DeleteCategoryResult{}, childErr
	}
	if hasChildren {
		now := time.Now().UTC()
		category.ArchivedAt = &now
		updated, _, saveErr := s.repo.SaveCategoryByIDForUser(normalizedUserID, normalizedCategoryID, category)
		if saveErr != nil {
			return DeleteCategoryResult{}, saveErr
		}
		return DeleteCategoryResult{Deleted: true, Archived: true, Category: updated}, &contractError{code: CAT_IN_USE_ARCHIVED}
	}

	referenced, refErr := s.repo.IsCategoryReferenced(normalizedUserID, normalizedCategoryID)
	if refErr != nil {
		return DeleteCategoryResult{}, refErr
	}
	if referenced {
		now := time.Now().UTC()
		category.ArchivedAt = &now
		updated, _, saveErr := s.repo.SaveCategoryByIDForUser(normalizedUserID, normalizedCategoryID, category)
		if saveErr != nil {
			return DeleteCategoryResult{}, saveErr
		}
		return DeleteCategoryResult{Deleted: true, Archived: true, Category: updated}, &contractError{code: CAT_IN_USE_ARCHIVED}
	}

	deleted, deleteErr := s.repo.DeleteCategoryByIDForUser(normalizedUserID, normalizedCategoryID)
	if deleteErr != nil {
		return DeleteCategoryResult{}, deleteErr
	}
	if !deleted {
		return DeleteCategoryResult{}, &contractError{code: CAT_NOT_FOUND}
	}
	return DeleteCategoryResult{Deleted: true, Archived: false, Category: category}, nil
}

func (s *CategoryService) ValidateCategorySelectable(_ context.Context, userID string, categoryID string) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCategoryID := strings.TrimSpace(categoryID)
	if normalizedUserID == "" || normalizedCategoryID == "" {
		return &contractError{code: CAT_INVALID}
	}

	category, found, err := s.repo.GetCategoryByIDForUser(normalizedUserID, normalizedCategoryID)
	if err != nil {
		return err
	}
	if !found {
		return &contractError{code: CAT_NOT_FOUND}
	}
	if category.ArchivedAt != nil {
		return &contractError{code: CAT_ARCHIVED}
	}
	return nil
}

func (s *CategoryService) RecordCategoryUsage(_ context.Context, userID string, categoryID string) (string, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCategoryID := strings.TrimSpace(categoryID)
	if normalizedUserID == "" || normalizedCategoryID == "" {
		return "", &contractError{code: CAT_INVALID}
	}

	category, found, err := s.repo.GetCategoryByIDForUser(normalizedUserID, normalizedCategoryID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", &contractError{code: CAT_NOT_FOUND}
	}
	if category.ArchivedAt != nil {
		return "", &contractError{code: CAT_ARCHIVED}
	}

	name, recErr := s.repo.RecordCategoryUsage(normalizedUserID, normalizedCategoryID)
	if recErr != nil {
		return "", recErr
	}
	if name == "" {
		return category.Name, nil
	}
	return name, nil
}

func (s *CategoryService) GetHistoricalCategoryName(_ context.Context, userID string, categoryID string) (string, bool) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCategoryID := strings.TrimSpace(categoryID)
	if normalizedUserID == "" || normalizedCategoryID == "" {
		return "", false
	}
	return s.repo.GetHistoricalCategoryName(normalizedUserID, normalizedCategoryID)
}

func (s *CategoryService) validateParent(userID string, categoryID string, parentID *string) error {
	normalizedParentID := strings.TrimSpace(ptrString(parentID))
	if normalizedParentID == "" {
		return nil
	}
	if normalizedParentID == strings.TrimSpace(categoryID) {
		return &contractError{code: CAT_INVALID_PARENT}
	}

	parent, found, err := s.repo.GetCategoryByIDForUser(userID, normalizedParentID)
	if err != nil {
		return err
	}
	if !found {
		return &contractError{code: CAT_INVALID_PARENT}
	}
	if parent.ArchivedAt != nil {
		return &contractError{code: CAT_INVALID_PARENT}
	}
	if strings.TrimSpace(ptrString(parent.ParentID)) != "" {
		return &contractError{code: CAT_INVALID_PARENT}
	}
	return nil
}

func normalizeOptionalID(value *string) *string {
	trimmed := strings.TrimSpace(ptrString(value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
