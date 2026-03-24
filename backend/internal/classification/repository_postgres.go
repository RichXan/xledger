package classification

import (
	"database/sql"
	"errors"
	"strings"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateCategory(userID string, input CategoryCreateInput) (Category, error) {
	var category Category
	var parentID *string
	if input.ParentID != nil && strings.TrimSpace(*input.ParentID) != "" {
		parentID = input.ParentID
	}
	err := r.db.QueryRow(`
		INSERT INTO categories (id, user_id, name, parent_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		RETURNING id, user_id, name, parent_id, archived_at
	`, userID, input.Name, parentID).Scan(&category.ID, &category.UserID, &category.Name, &category.ParentID, &category.ArchivedAt)
	return category, err
}

func (r *PostgresRepository) ListCategoriesByUser(userID string) ([]Category, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, parent_id, archived_at
		FROM categories
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Category, 0)
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.UserID, &category.Name, &category.ParentID, &category.ArchivedAt); err != nil {
			return nil, err
		}
		items = append(items, category)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetCategoryByIDForUser(userID string, categoryID string) (Category, bool, error) {
	var category Category
	err := r.db.QueryRow(`
		SELECT id, user_id, name, parent_id, archived_at
		FROM categories
		WHERE id = $1 AND user_id = $2
	`, categoryID, userID).Scan(&category.ID, &category.UserID, &category.Name, &category.ParentID, &category.ArchivedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Category{}, false, nil
	}
	if err != nil {
		return Category{}, false, err
	}
	return category, true, nil
}

func (r *PostgresRepository) SaveCategoryByIDForUser(userID string, categoryID string, category Category) (Category, bool, error) {
	var updated Category
	err := r.db.QueryRow(`
		UPDATE categories
		SET name = $3, parent_id = $4, archived_at = $5
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, parent_id, archived_at
	`, categoryID, userID, category.Name, category.ParentID, category.ArchivedAt).Scan(
		&updated.ID, &updated.UserID, &updated.Name, &updated.ParentID, &updated.ArchivedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Category{}, false, nil
	}
	if err != nil {
		return Category{}, false, err
	}
	return updated, true, nil
}

func (r *PostgresRepository) DeleteCategoryByIDForUser(userID string, categoryID string) (bool, error) {
	result, err := r.db.Exec(`
		DELETE FROM categories
		WHERE id = $1 AND user_id = $2
	`, categoryID, userID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *PostgresRepository) CategoryHasChildren(userID string, categoryID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM categories
			WHERE parent_id = $1 AND user_id = $2
		)
	`, categoryID, userID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) IsCategoryReferenced(userID string, categoryID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM transactions
		WHERE category_id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, categoryID, userID).Scan(&count)
	return count > 0, err
}

func (r *PostgresRepository) RecordCategoryUsage(userID string, categoryID string) (string, error) {
	var name string
	err := r.db.QueryRow(`
		UPDATE categories
		SET usage_count = COALESCE(usage_count, 0) + 1
		WHERE id = $1 AND user_id = $2
		RETURNING name
	`, categoryID, userID).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return name, err
}

func (r *PostgresRepository) GetHistoricalCategoryName(userID string, categoryID string) (string, bool) {
	var name string
	err := r.db.QueryRow(`
		SELECT name FROM category_history
		WHERE category_id = $1 AND user_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, categoryID, userID).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return name, true
}

func (r *PostgresRepository) CreateTag(userID string, input TagCreateInput) (Tag, error) {
	var tag Tag
	err := r.db.QueryRow(`
		INSERT INTO tags (id, user_id, name, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		RETURNING id, user_id, name
	`, userID, input.Name).Scan(&tag.ID, &tag.UserID, &tag.Name)
	return tag, err
}

func (r *PostgresRepository) ListTagsByUser(userID string) ([]Tag, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name
		FROM tags
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Tag, 0)
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name); err != nil {
			return nil, err
		}
		items = append(items, tag)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetTagByIDForUser(userID string, tagID string) (Tag, bool, error) {
	var tag Tag
	err := r.db.QueryRow(`
		SELECT id, user_id, name
		FROM tags
		WHERE id = $1 AND user_id = $2
	`, tagID, userID).Scan(&tag.ID, &tag.UserID, &tag.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return Tag{}, false, nil
	}
	if err != nil {
		return Tag{}, false, err
	}
	return tag, true, nil
}

func (r *PostgresRepository) GetTagByNameForUser(userID string, normalizedName string) (Tag, bool, error) {
	var tag Tag
	err := r.db.QueryRow(`
		SELECT id, user_id, name
		FROM tags
		WHERE user_id = $1 AND LOWER(TRIM(name)) = LOWER(TRIM($2))
	`, userID, normalizedName).Scan(&tag.ID, &tag.UserID, &tag.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return Tag{}, false, nil
	}
	if err != nil {
		return Tag{}, false, err
	}
	return tag, true, nil
}

func (r *PostgresRepository) SaveTagByIDForUser(userID string, tagID string, tag Tag) (Tag, bool, error) {
	var updated Tag
	err := r.db.QueryRow(`
		UPDATE tags
		SET name = $3
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name
	`, tagID, userID, tag.Name).Scan(&updated.ID, &updated.UserID, &updated.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return Tag{}, false, nil
	}
	if err != nil {
		return Tag{}, false, err
	}
	return updated, true, nil
}

func (r *PostgresRepository) DeleteTagByIDForUser(userID string, tagID string) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM transaction_tags
		WHERE tag_id = $1 AND user_id = $2
	`, tagID, userID)
	if err != nil {
		return false, err
	}

	result, err := tx.Exec(`
		DELETE FROM tags
		WHERE id = $1 AND user_id = $2
	`, tagID, userID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, tx.Commit()
}

func (r *PostgresRepository) AttachTagToTransaction(userID string, tagID string, transactionID string) error {
	_, err := r.db.Exec(`
		INSERT INTO transaction_tags (transaction_id, tag_id, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, transactionID, tagID, userID)
	return err
}

func (r *PostgresRepository) ReplaceTransactionTags(userID string, transactionID string, tagIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM transaction_tags
		WHERE transaction_id = $1 AND user_id = $2
	`, transactionID, userID)
	if err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		_, err = tx.Exec(`
			INSERT INTO transaction_tags (transaction_id, tag_id, user_id)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
		`, transactionID, tagID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) RemoveTransactionTags(userID string, transactionID string) error {
	_, err := r.db.Exec(`
		DELETE FROM transaction_tags
		WHERE transaction_id = $1 AND user_id = $2
	`, transactionID, userID)
	return err
}

func (r *PostgresRepository) ListTransactionIDsByTag(userID string, tagID string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT transaction_id
		FROM transaction_tags
		WHERE tag_id = $1 AND user_id = $2
	`, tagID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0)
	for rows.Next() {
		var txnID string
		if err := rows.Scan(&txnID); err != nil {
			return nil, err
		}
		items = append(items, txnID)
	}
	return items, rows.Err()
}
