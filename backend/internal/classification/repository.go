package classification

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Category struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	ParentID   *string    `json:"parent_id,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

type CategoryCreateInput struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
}

type CategoryUpdateInput struct {
	Name        *string `json:"name,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
	ClearParent bool    `json:"clear_parent,omitempty"`
	Archive     *bool   `json:"archived,omitempty"`
}

type Tag struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

type TagCreateInput struct {
	Name string `json:"name"`
}

type TagUpdateInput struct {
	Name string `json:"name"`
}

type Repository interface {
	CreateCategory(userID string, input CategoryCreateInput) (Category, error)
	ListCategoriesByUser(userID string) ([]Category, error)
	GetCategoryByIDForUser(userID string, categoryID string) (Category, bool, error)
	SaveCategoryByIDForUser(userID string, categoryID string, category Category) (Category, bool, error)
	DeleteCategoryByIDForUser(userID string, categoryID string) (bool, error)
	CategoryHasChildren(userID string, categoryID string) (bool, error)
	IsCategoryReferenced(userID string, categoryID string) (bool, error)
	RecordCategoryUsage(userID string, categoryID string) (string, error)
	GetHistoricalCategoryName(userID string, categoryID string) (string, bool)

	CreateTag(userID string, input TagCreateInput) (Tag, error)
	ListTagsByUser(userID string) ([]Tag, error)
	GetTagByIDForUser(userID string, tagID string) (Tag, bool, error)
	GetTagByNameForUser(userID string, normalizedName string) (Tag, bool, error)
	SaveTagByIDForUser(userID string, tagID string, tag Tag) (Tag, bool, error)
	DeleteTagByIDForUser(userID string, tagID string) (bool, error)
	AttachTagToTransaction(userID string, tagID string, transactionID string) error
	ReplaceTransactionTags(userID string, transactionID string, tagIDs []string) error
	RemoveTransactionTags(userID string, transactionID string) error
	ListTransactionIDsByTag(userID string, tagID string) ([]string, error)
}

type InMemoryRepository struct {
	mu sync.Mutex

	categories map[string]Category
	tags       map[string]Tag

	categoryReferenceCount map[string]int
	categoryHistoryNames   map[string]string

	tagTransactionOrder map[string][]string
	tagTransactionSet   map[string]map[string]bool
}

var classificationIDCounter int64

func nextClassificationID() string {
	value := atomic.AddInt64(&classificationIDCounter, 1)
	return "id-" + strconv.FormatInt(value, 10)
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		categories:             map[string]Category{},
		tags:                   map[string]Tag{},
		categoryReferenceCount: map[string]int{},
		categoryHistoryNames:   map[string]string{},
		tagTransactionOrder:    map[string][]string{},
		tagTransactionSet:      map[string]map[string]bool{},
	}
}

func (r *InMemoryRepository) CreateCategory(userID string, input CategoryCreateInput) (Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	category := Category{
		ID:       nextClassificationID(),
		UserID:   userID,
		Name:     input.Name,
		ParentID: cloneStringPtr(input.ParentID),
	}
	r.categories[category.ID] = category
	return category, nil
}

func (r *InMemoryRepository) ListCategoriesByUser(userID string) ([]Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]Category, 0)
	for _, category := range r.categories {
		if category.UserID != userID {
			continue
		}
		items = append(items, category)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (r *InMemoryRepository) GetCategoryByIDForUser(userID string, categoryID string) (Category, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	category, ok := r.categories[categoryID]
	if !ok || category.UserID != userID {
		return Category{}, false, nil
	}
	return category, true, nil
}

func (r *InMemoryRepository) SaveCategoryByIDForUser(userID string, categoryID string, category Category) (Category, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.categories[categoryID]
	if !ok || current.UserID != userID {
		return Category{}, false, nil
	}
	category.ID = current.ID
	category.UserID = current.UserID
	r.categories[category.ID] = category
	return category, true, nil
}

func (r *InMemoryRepository) DeleteCategoryByIDForUser(userID string, categoryID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	category, ok := r.categories[categoryID]
	if !ok || category.UserID != userID {
		return false, nil
	}
	delete(r.categories, categoryID)
	return true, nil
}

func (r *InMemoryRepository) CategoryHasChildren(userID string, categoryID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, category := range r.categories {
		if category.UserID != userID {
			continue
		}
		if strings.TrimSpace(ptrString(category.ParentID)) != categoryID {
			continue
		}
		return true, nil
	}
	return false, nil
}

func (r *InMemoryRepository) IsCategoryReferenced(userID string, categoryID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.categoryReferenceCount[userID+"|"+categoryID] > 0, nil
}

func (r *InMemoryRepository) RecordCategoryUsage(userID string, categoryID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	category, ok := r.categories[categoryID]
	if !ok || category.UserID != userID {
		return "", nil
	}
	refKey := userID + "|" + categoryID
	r.categoryReferenceCount[refKey]++
	r.categoryHistoryNames[refKey] = category.Name
	return category.Name, nil
}

func (r *InMemoryRepository) GetHistoricalCategoryName(userID string, categoryID string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name, ok := r.categoryHistoryNames[userID+"|"+categoryID]
	return name, ok
}

func (r *InMemoryRepository) CreateTag(userID string, input TagCreateInput) (Tag, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tag := Tag{ID: nextClassificationID(), UserID: userID, Name: input.Name}
	r.tags[tag.ID] = tag
	return tag, nil
}

func (r *InMemoryRepository) ListTagsByUser(userID string) ([]Tag, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]Tag, 0)
	for _, tag := range r.tags {
		if tag.UserID != userID {
			continue
		}
		items = append(items, tag)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (r *InMemoryRepository) GetTagByIDForUser(userID string, tagID string) (Tag, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tag, ok := r.tags[tagID]
	if !ok || tag.UserID != userID {
		return Tag{}, false, nil
	}
	return tag, true, nil
}

func (r *InMemoryRepository) GetTagByNameForUser(userID string, normalizedName string) (Tag, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	needle := strings.ToLower(strings.TrimSpace(normalizedName))
	for _, tag := range r.tags {
		if tag.UserID != userID {
			continue
		}
		if strings.ToLower(strings.TrimSpace(tag.Name)) == needle {
			return tag, true, nil
		}
	}
	return Tag{}, false, nil
}

func (r *InMemoryRepository) SaveTagByIDForUser(userID string, tagID string, tag Tag) (Tag, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.tags[tagID]
	if !ok || current.UserID != userID {
		return Tag{}, false, nil
	}
	tag.ID = current.ID
	tag.UserID = current.UserID
	r.tags[tag.ID] = tag
	return tag, true, nil
}

func (r *InMemoryRepository) DeleteTagByIDForUser(userID string, tagID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tag, ok := r.tags[tagID]
	if !ok || tag.UserID != userID {
		return false, nil
	}
	delete(r.tags, tagID)
	delete(r.tagTransactionOrder, userID+"|"+tagID)
	delete(r.tagTransactionSet, userID+"|"+tagID)
	return true, nil
}

func (r *InMemoryRepository) AttachTagToTransaction(userID string, tagID string, transactionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tag, ok := r.tags[tagID]
	if !ok || tag.UserID != userID {
		return nil
	}
	key := userID + "|" + tagID
	if r.tagTransactionSet[key] == nil {
		r.tagTransactionSet[key] = map[string]bool{}
	}
	if r.tagTransactionSet[key][transactionID] {
		return nil
	}
	r.tagTransactionSet[key][transactionID] = true
	r.tagTransactionOrder[key] = append(r.tagTransactionOrder[key], transactionID)
	return nil
}

func (r *InMemoryRepository) ReplaceTransactionTags(userID string, transactionID string, tagIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.removeTransactionFromAllTagsLocked(userID, transactionID)
	for _, tagID := range tagIDs {
		tag, ok := r.tags[tagID]
		if !ok || tag.UserID != userID {
			continue
		}
		r.attachTagToTransactionLocked(userID, tagID, transactionID)
	}
	return nil
}

func (r *InMemoryRepository) RemoveTransactionTags(userID string, transactionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.removeTransactionFromAllTagsLocked(userID, transactionID)
	return nil
}

func (r *InMemoryRepository) ListTransactionIDsByTag(userID string, tagID string) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.tags[tagID]
	if !ok {
		return []string{}, nil
	}
	key := userID + "|" + tagID
	items := make([]string, 0, len(r.tagTransactionOrder[key]))
	items = append(items, r.tagTransactionOrder[key]...)
	return items, nil
}

func (r *InMemoryRepository) attachTagToTransactionLocked(userID string, tagID string, transactionID string) {
	key := userID + "|" + tagID
	if r.tagTransactionSet[key] == nil {
		r.tagTransactionSet[key] = map[string]bool{}
	}
	if r.tagTransactionSet[key][transactionID] {
		return
	}
	r.tagTransactionSet[key][transactionID] = true
	r.tagTransactionOrder[key] = append(r.tagTransactionOrder[key], transactionID)
}

func (r *InMemoryRepository) removeTransactionFromAllTagsLocked(userID string, transactionID string) {
	for key, set := range r.tagTransactionSet {
		if !strings.HasPrefix(key, userID+"|") {
			continue
		}
		if !set[transactionID] {
			continue
		}
		delete(set, transactionID)
		order := r.tagTransactionOrder[key]
		filtered := order[:0]
		for _, existing := range order {
			if existing == transactionID {
				continue
			}
			filtered = append(filtered, existing)
		}
		if len(filtered) == 0 {
			delete(r.tagTransactionOrder, key)
			delete(r.tagTransactionSet, key)
			continue
		}
		r.tagTransactionOrder[key] = filtered
	}
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
