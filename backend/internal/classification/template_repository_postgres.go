package classification

import (
	"context"
	"database/sql"
)

type PostgresTemplateRepository struct {
	db *sql.DB
}

func NewPostgresTemplateRepository(db *sql.DB) *PostgresTemplateRepository {
	return &PostgresTemplateRepository{db: db}
}

func (r *PostgresTemplateRepository) CopyDefaultTemplateToUser(ctx context.Context, userID string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_category_templates WHERE user_id = $1)
	`, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, parent_id FROM default_categories ORDER BY parent_id NULLS FIRST, sort_order, name
	`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	type defaultCategory struct {
		DefaultID string
		Name      string
		ParentID  *string
	}
	var defaults []defaultCategory
	for rows.Next() {
		var dc defaultCategory
		if err := rows.Scan(&dc.DefaultID, &dc.Name, &dc.ParentID); err != nil {
			return false, err
		}
		defaults = append(defaults, dc)
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	idMapping := make(map[string]string)
	for _, dc := range defaults {
		var newID string
		var parentID *string
		if dc.ParentID != nil {
			if mapped, ok := idMapping[*dc.ParentID]; ok {
				parentID = &mapped
			}
		}
		err = tx.QueryRowContext(ctx, `
			INSERT INTO categories (id, user_id, name, parent_id, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, NOW())
			RETURNING id
		`, userID, dc.Name, parentID).Scan(&newID)
		if err != nil {
			return false, err
		}
		idMapping[dc.DefaultID] = newID
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_category_templates (user_id, copied_at)
		VALUES ($1, NOW())
	`, userID)
	if err != nil {
		return false, err
	}

	return true, tx.Commit()
}

type UserTemplateRepository struct {
	db *sql.DB
}

func NewUserTemplateRepository(db *sql.DB) *UserTemplateRepository {
	return &UserTemplateRepository{db: db}
}

func (r *UserTemplateRepository) HasUserTemplate(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_category_templates WHERE user_id = $1)
	`, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
