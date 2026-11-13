SELECT city, COUNT(*) AS user_count
FROM users
GROUP BY city
HAVING COUNT(*) > 1
ORDER BY user_count DESC, city ASC;
