CREATE TABLE IF NOT EXISTS default_categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_category_templates (
    user_id UUID PRIMARY KEY,
    copied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO default_categories (id, name, parent_id, sort_order) VALUES
    ('food', '餐饮', NULL, 1),
    ('food_breakfast', '早餐', 'food', 1),
    ('food_lunch', '午餐', 'food', 2),
    ('food_dinner', '晚餐', 'food', 3),
    ('food_snacks', '零食', 'food', 4),
    ('transport', '交通', NULL, 2),
    ('transport_public', '公共交通', 'transport', 1),
    ('transport_taxi', '出租车/网约车', 'transport', 2),
    ('transport_fuel', '燃油', 'transport', 3),
    ('shopping', '购物', NULL, 3),
    ('shopping_clothes', '服装', 'shopping', 1),
    ('shopping_daily', '日用品', 'shopping', 2),
    ('shopping_electronics', '数码产品', 'shopping', 3),
    ('housing', '住房', NULL, 4),
    ('housing_rent', '房租', 'housing', 1),
    ('housing_utilities', '水电煤', 'housing', 2),
    ('housing_property', '物业费', 'housing', 3),
    ('entertainment', '娱乐', NULL, 5),
    ('entertainment_games', '游戏', 'entertainment', 1),
    ('entertainment_movies', '电影', 'entertainment', 2),
    ('entertainment_travel', '旅行', 'entertainment', 3),
    ('health', '医疗健康', NULL, 6),
    ('health_medical', '看病', 'health', 1),
    ('health_pharmacy', '药品', 'health', 2),
    ('health_fitness', '健身', 'health', 3),
    ('education', '教育', NULL, 7),
    ('education_books', '书籍', 'education', 1),
    ('education_courses', '课程', 'education', 2),
    ('income', '收入', NULL, 8),
    ('income_salary', '工资', 'income', 1),
    ('income_bonus', '奖金', 'income', 2),
    ('income_investment', '投资收益', 'income', 3),
    ('income_other', '其他收入', 'income', 4),
    ('other', '其他', NULL, 99)
ON CONFLICT (id) DO NOTHING;
