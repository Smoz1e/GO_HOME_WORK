SELECT u.name, COALESCE(SUM(o.price), 0) AS total_spent
FROM users AS u
LEFT JOIN orders AS o ON o.user_id = u.id
GROUP BY u.id, u.name
ORDER BY total_spent DESC, u.name ASC;
