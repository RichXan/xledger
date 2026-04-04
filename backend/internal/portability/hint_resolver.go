package portability

import (
    "context"
    "strings"

    "xledger/backend/internal/classification"
    "xledger/backend/internal/accounting"
)

type HintResolver struct {
    categoryService *classification.CategoryService
    accountRepo     accounting.AccountRepository
}

func NewHintResolver(categoryService *classification.CategoryService, accountRepo accounting.AccountRepository) *HintResolver {
    return &HintResolver{
        categoryService: categoryService,
        accountRepo:     accountRepo,
    }
}

type ResolvedHints struct {
    CategoryID   *string
    AccountID   *string
    Description string
}

// ResolveCategoryHint fuzzy-matches a category name to a user category ID.
// Returns nil if no match or multiple matches (ambiguous).
func (r *HintResolver) ResolveCategoryHint(ctx context.Context, userID, hint string) (*string, error) {
    if hint == "" {
        return nil, nil
    }

    hint = strings.TrimSpace(hint)
    categories, err := r.categoryService.ListCategories(ctx, userID)
    if err != nil {
        return nil, err
    }

    var matches []string
    hintLower := strings.ToLower(hint)

    for _, cat := range categories {
        nameLower := strings.ToLower(cat.Name)
        // Exact contains match
        if strings.Contains(nameLower, hintLower) || strings.Contains(hintLower, nameLower) {
            matches = append(matches, cat.ID)
            continue
        }
        // Keyword match
        keywords := strings.Fields(hintLower)
        allMatch := true
        for _, kw := range keywords {
            if !strings.Contains(nameLower, kw) {
                allMatch = false
                break
            }
        }
        if allMatch {
            matches = append(matches, cat.ID)
        }
    }

    if len(matches) == 1 {
        return &matches[0], nil
    }
    // Multiple matches (ambiguous) or no match: return nil
    return nil, nil
}

// ResolveAccountHint fuzzy-matches an account name to a user account ID.
func (r *HintResolver) ResolveAccountHint(ctx context.Context, userID, hint string) (*string, error) {
    if hint == "" {
        return nil, nil
    }

    hint = strings.TrimSpace(hint)
    accounts, err := r.accountRepo.ListByUser(userID)
    if err != nil {
        return nil, err
    }

    var matches []string
    hintLower := strings.ToLower(hint)

    for _, acct := range accounts {
        nameLower := strings.ToLower(acct.Name)
        if strings.Contains(nameLower, hintLower) || strings.Contains(hintLower, nameLower) {
            matches = append(matches, acct.ID)
        }
    }

    if len(matches) == 1 {
        return &matches[0], nil
    }
    return nil, nil
}
