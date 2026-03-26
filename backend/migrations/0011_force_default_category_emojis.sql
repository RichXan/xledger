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
