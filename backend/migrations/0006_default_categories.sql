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
    ('food', 'Food', NULL, 1),
    ('food_breakfast', 'Breakfast', 'food', 1),
    ('food_lunch', 'Lunch', 'food', 2),
    ('food_dinner', 'Dinner', 'food', 3),
    ('food_snacks', 'Snacks', 'food', 4),
    ('transport', 'Transport', NULL, 2),
    ('transport_public', 'Public Transit', 'transport', 1),
    ('transport_taxi', 'Taxi and Ride-hailing', 'transport', 2),
    ('transport_fuel', 'Fuel', 'transport', 3),
    ('shopping', 'Shopping', NULL, 3),
    ('shopping_clothes', 'Clothes', 'shopping', 1),
    ('shopping_daily', 'Daily Needs', 'shopping', 2),
    ('shopping_electronics', 'Electronics', 'shopping', 3),
    ('housing', 'Housing', NULL, 4),
    ('housing_rent', 'Rent', 'housing', 1),
    ('housing_utilities', 'Utilities', 'housing', 2),
    ('housing_property', 'Property Fees', 'housing', 3),
    ('entertainment', 'Entertainment', NULL, 5),
    ('entertainment_games', 'Games', 'entertainment', 1),
    ('entertainment_movies', 'Movies', 'entertainment', 2),
    ('entertainment_travel', 'Travel', 'entertainment', 3),
    ('health', 'Health', NULL, 6),
    ('health_medical', 'Medical', 'health', 1),
    ('health_pharmacy', 'Pharmacy', 'health', 2),
    ('health_fitness', 'Fitness', 'health', 3),
    ('education', 'Education', NULL, 7),
    ('education_books', 'Books', 'education', 1),
    ('education_courses', 'Courses', 'education', 2),
    ('income', 'Income', NULL, 8),
    ('income_salary', 'Salary', 'income', 1),
    ('income_bonus', 'Bonus', 'income', 2),
    ('income_investment', 'Investment Income', 'income', 3),
    ('income_other', 'Other Income', 'income', 4),
    ('other', 'Other', NULL, 99)
ON CONFLICT (id) DO NOTHING;
