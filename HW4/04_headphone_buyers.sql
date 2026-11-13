SELECT DISTINCT u.name
FROM users AS u
JOIN orders AS o ON o.user_id = u.id
WHERE o.item = 'наушники'
ORDER BY u.name;
