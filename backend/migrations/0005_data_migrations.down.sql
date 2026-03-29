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
SET name = m.plain_name
FROM mapping m
WHERE c.name = m.emoji_name;

DELETE FROM default_categories
WHERE id IN (
    'food',
    'food_breakfast',
    'food_lunch',
    'food_dinner',
    'food_snacks',
    'transport',
    'transport_public',
    'transport_taxi',
    'transport_fuel',
    'shopping',
    'shopping_clothes',
    'shopping_daily',
    'shopping_electronics',
    'housing',
    'housing_rent',
    'housing_utilities',
    'housing_property',
    'entertainment',
    'entertainment_games',
    'entertainment_movies',
    'entertainment_travel',
    'health',
    'health_medical',
    'health_pharmacy',
    'health_fitness',
    'education',
    'education_books',
    'education_courses',
    'income',
    'income_salary',
    'income_bonus',
    'income_investment',
    'income_other',
    'other'
);
