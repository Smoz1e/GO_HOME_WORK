SELECT name, age
FROM users
WHERE city = 'Москва'
  AND age = (
      SELECT MAX(age)
      FROM users
      WHERE city = 'Москва'
  )
ORDER BY name;
