INSERT INTO default_categories (id, name, parent_id, sort_order) VALUES
    ('food', '🍱 Food', NULL, 1),
    ('food_breakfast', '🥐 Breakfast', 'food', 1),
    ('food_lunch', '🍜 Lunch', 'food', 2),
    ('food_dinner', '🍲 Dinner', 'food', 3),
    ('food_snacks', '🍿 Snacks', 'food', 4),
    ('transport', '🚕 Transport', NULL, 2),
    ('transport_public', '🚌 Public Transit', 'transport', 1),
    ('transport_taxi', '🚖 Taxi and Ride-hailing', 'transport', 2),
    ('transport_fuel', '⛽ Fuel', 'transport', 3),
    ('shopping', '🛍️ Shopping', NULL, 3),
    ('shopping_clothes', '👕 Clothes', 'shopping', 1),
    ('shopping_daily', '🧻 Daily Needs', 'shopping', 2),
    ('shopping_electronics', '💻 Electronics', 'shopping', 3),
    ('housing', '🏠 Housing', NULL, 4),
    ('housing_rent', '🏡 Rent', 'housing', 1),
    ('housing_utilities', '💡 Utilities', 'housing', 2),
    ('housing_property', '🧾 Property Fees', 'housing', 3),
    ('entertainment', '🎮 Entertainment', NULL, 5),
    ('entertainment_games', '🕹️ Games', 'entertainment', 1),
    ('entertainment_movies', '🎬 Movies', 'entertainment', 2),
    ('entertainment_travel', '✈️ Travel', 'entertainment', 3),
    ('health', '🩺 Health', NULL, 6),
    ('health_medical', '🏥 Medical', 'health', 1),
    ('health_pharmacy', '💊 Pharmacy', 'health', 2),
    ('health_fitness', '💪 Fitness', 'health', 3),
    ('education', '📚 Education', NULL, 7),
    ('education_books', '📖 Books', 'education', 1),
    ('education_courses', '🧠 Courses', 'education', 2),
    ('income', '💰 Income', NULL, 8),
    ('income_salary', '💼 Salary', 'income', 1),
    ('income_bonus', '🎁 Bonus', 'income', 2),
    ('income_investment', '📈 Investment Income', 'income', 3),
    ('income_other', '🪙 Other Income', 'income', 4),
    ('other', '📦 Other', NULL, 99)
ON CONFLICT (id) DO NOTHING;

WITH mapping AS (
    SELECT *
    FROM (
        VALUES
            ('food', '🍱 Food', 'Food'),
            ('food_breakfast', '🥐 Breakfast', 'Breakfast'),
            ('food_lunch', '🍜 Lunch', 'Lunch'),
            ('food_dinner', '🍲 Dinner', 'Dinner'),
            ('food_snacks', '🍿 Snacks', 'Snacks'),
            ('transport', '🚕 Transport', 'Transport'),
            ('transport_public', '🚌 Public Transit', 'Public Transit'),
            ('transport_taxi', '🚖 Taxi and Ride-hailing', 'Taxi and Ride-hailing'),
            ('transport_fuel', '⛽ Fuel', 'Fuel'),
            ('shopping', '🛍️ Shopping', 'Shopping'),
            ('shopping_clothes', '👕 Clothes', 'Clothes'),
            ('shopping_daily', '🧻 Daily Needs', 'Daily Needs'),
            ('shopping_electronics', '💻 Electronics', 'Electronics'),
            ('housing', '🏠 Housing', 'Housing'),
            ('housing_rent', '🏡 Rent', 'Rent'),
            ('housing_utilities', '💡 Utilities', 'Utilities'),
            ('housing_property', '🧾 Property Fees', 'Property Fees'),
            ('entertainment', '🎮 Entertainment', 'Entertainment'),
            ('entertainment_games', '🕹️ Games', 'Games'),
            ('entertainment_movies', '🎬 Movies', 'Movies'),
            ('entertainment_travel', '✈️ Travel', 'Travel'),
            ('health', '🩺 Health', 'Health'),
            ('health_medical', '🏥 Medical', 'Medical'),
            ('health_pharmacy', '💊 Pharmacy', 'Pharmacy'),
            ('health_fitness', '💪 Fitness', 'Fitness'),
            ('education', '📚 Education', 'Education'),
            ('education_books', '📖 Books', 'Books'),
            ('education_courses', '🧠 Courses', 'Courses'),
            ('income', '💰 Income', 'Income'),
            ('income_salary', '💼 Salary', 'Salary'),
            ('income_bonus', '🎁 Bonus', 'Bonus'),
            ('income_investment', '📈 Investment Income', 'Investment Income'),
            ('income_other', '🪙 Other Income', 'Other Income'),
            ('other', '📦 Other', 'Other')
    ) AS t(id, emoji_name, plain_name)
)
UPDATE default_categories d
SET name = m.emoji_name
FROM mapping m
WHERE d.id = m.id;

WITH mapping AS (
    SELECT *
    FROM (
        VALUES
            ('🍱 Food', 'Food'),
            ('🥐 Breakfast', 'Breakfast'),
            ('🍜 Lunch', 'Lunch'),
            ('🍲 Dinner', 'Dinner'),
            ('🍿 Snacks', 'Snacks'),
            ('🚕 Transport', 'Transport'),
            ('🚌 Public Transit', 'Public Transit'),
            ('🚖 Taxi and Ride-hailing', 'Taxi and Ride-hailing'),
            ('⛽ Fuel', 'Fuel'),
            ('🛍️ Shopping', 'Shopping'),
            ('👕 Clothes', 'Clothes'),
            ('🧻 Daily Needs', 'Daily Needs'),
            ('💻 Electronics', 'Electronics'),
            ('🏠 Housing', 'Housing'),
            ('🏡 Rent', 'Rent'),
            ('💡 Utilities', 'Utilities'),
            ('🧾 Property Fees', 'Property Fees'),
            ('🎮 Entertainment', 'Entertainment'),
            ('🕹️ Games', 'Games'),
            ('🎬 Movies', 'Movies'),
            ('✈️ Travel', 'Travel'),
            ('🩺 Health', 'Health'),
            ('🏥 Medical', 'Medical'),
            ('💊 Pharmacy', 'Pharmacy'),
            ('💪 Fitness', 'Fitness'),
            ('📚 Education', 'Education'),
            ('📖 Books', 'Books'),
            ('🧠 Courses', 'Courses'),
            ('💰 Income', 'Income'),
            ('💼 Salary', 'Salary'),
            ('🎁 Bonus', 'Bonus'),
            ('📈 Investment Income', 'Investment Income'),
            ('🪙 Other Income', 'Other Income'),
            ('📦 Other', 'Other')
    ) AS t(emoji_name, plain_name)
)
UPDATE categories c
SET name = m.emoji_name
FROM mapping m
WHERE c.name = m.plain_name;
