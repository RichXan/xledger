UPDATE default_categories
SET name = CASE id
    WHEN 'food' THEN '🍱 Food'
    WHEN 'food_breakfast' THEN '🥐 Breakfast'
    WHEN 'food_lunch' THEN '🍜 Lunch'
    WHEN 'food_dinner' THEN '🍲 Dinner'
    WHEN 'food_snacks' THEN '🍿 Snacks'
    WHEN 'transport' THEN '🚕 Transport'
    WHEN 'transport_public' THEN '🚌 Public Transit'
    WHEN 'transport_taxi' THEN '🚖 Taxi and Ride-hailing'
    WHEN 'transport_fuel' THEN '⛽ Fuel'
    WHEN 'shopping' THEN '🛍️ Shopping'
    WHEN 'shopping_clothes' THEN '👕 Clothes'
    WHEN 'shopping_daily' THEN '🧻 Daily Needs'
    WHEN 'shopping_electronics' THEN '💻 Electronics'
    WHEN 'housing' THEN '🏠 Housing'
    WHEN 'housing_rent' THEN '🏡 Rent'
    WHEN 'housing_utilities' THEN '💡 Utilities'
    WHEN 'housing_property' THEN '🧾 Property Fees'
    WHEN 'entertainment' THEN '🎮 Entertainment'
    WHEN 'entertainment_games' THEN '🕹️ Games'
    WHEN 'entertainment_movies' THEN '🎬 Movies'
    WHEN 'entertainment_travel' THEN '✈️ Travel'
    WHEN 'health' THEN '🩺 Health'
    WHEN 'health_medical' THEN '🏥 Medical'
    WHEN 'health_pharmacy' THEN '💊 Pharmacy'
    WHEN 'health_fitness' THEN '💪 Fitness'
    WHEN 'education' THEN '📚 Education'
    WHEN 'education_books' THEN '📖 Books'
    WHEN 'education_courses' THEN '🧠 Courses'
    WHEN 'income' THEN '💰 Income'
    WHEN 'income_salary' THEN '💼 Salary'
    WHEN 'income_bonus' THEN '🎁 Bonus'
    WHEN 'income_investment' THEN '📈 Investment Income'
    WHEN 'income_other' THEN '🪙 Other Income'
    WHEN 'other' THEN '📦 Other'
    ELSE name
END
WHERE id IN (
    'food','food_breakfast','food_lunch','food_dinner','food_snacks',
    'transport','transport_public','transport_taxi','transport_fuel',
    'shopping','shopping_clothes','shopping_daily','shopping_electronics',
    'housing','housing_rent','housing_utilities','housing_property',
    'entertainment','entertainment_games','entertainment_movies','entertainment_travel',
    'health','health_medical','health_pharmacy','health_fitness',
    'education','education_books','education_courses',
    'income','income_salary','income_bonus','income_investment','income_other',
    'other'
);

UPDATE categories
SET name = CASE name
    WHEN 'Food' THEN '🍱 Food'
    WHEN 'Breakfast' THEN '🥐 Breakfast'
    WHEN 'Lunch' THEN '🍜 Lunch'
    WHEN 'Dinner' THEN '🍲 Dinner'
    WHEN 'Snacks' THEN '🍿 Snacks'
    WHEN 'Transport' THEN '🚕 Transport'
    WHEN 'Public Transit' THEN '🚌 Public Transit'
    WHEN 'Taxi and Ride-hailing' THEN '🚖 Taxi and Ride-hailing'
    WHEN 'Fuel' THEN '⛽ Fuel'
    WHEN 'Shopping' THEN '🛍️ Shopping'
    WHEN 'Clothes' THEN '👕 Clothes'
    WHEN 'Daily Needs' THEN '🧻 Daily Needs'
    WHEN 'Electronics' THEN '💻 Electronics'
    WHEN 'Housing' THEN '🏠 Housing'
    WHEN 'Rent' THEN '🏡 Rent'
    WHEN 'Utilities' THEN '💡 Utilities'
    WHEN 'Property Fees' THEN '🧾 Property Fees'
    WHEN 'Entertainment' THEN '🎮 Entertainment'
    WHEN 'Games' THEN '🕹️ Games'
    WHEN 'Movies' THEN '🎬 Movies'
    WHEN 'Travel' THEN '✈️ Travel'
    WHEN 'Health' THEN '🩺 Health'
    WHEN 'Medical' THEN '🏥 Medical'
    WHEN 'Pharmacy' THEN '💊 Pharmacy'
    WHEN 'Fitness' THEN '💪 Fitness'
    WHEN 'Education' THEN '📚 Education'
    WHEN 'Books' THEN '📖 Books'
    WHEN 'Courses' THEN '🧠 Courses'
    WHEN 'Income' THEN '💰 Income'
    WHEN 'Salary' THEN '💼 Salary'
    WHEN 'Bonus' THEN '🎁 Bonus'
    WHEN 'Investment Income' THEN '📈 Investment Income'
    WHEN 'Other Income' THEN '🪙 Other Income'
    WHEN 'Other' THEN '📦 Other'
    ELSE name
END
WHERE name IN (
    'Food','Breakfast','Lunch','Dinner','Snacks',
    'Transport','Public Transit','Taxi and Ride-hailing','Fuel',
    'Shopping','Clothes','Daily Needs','Electronics',
    'Housing','Rent','Utilities','Property Fees',
    'Entertainment','Games','Movies','Travel',
    'Health','Medical','Pharmacy','Fitness',
    'Education','Books','Courses',
    'Income','Salary','Bonus','Investment Income','Other Income',
    'Other'
);
