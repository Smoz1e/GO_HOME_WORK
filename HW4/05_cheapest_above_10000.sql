SELECT o.item, o.price
FROM orders AS o
WHERE o.price > 10000
  AND o.price = (
      SELECT MIN(price)
      FROM orders
      WHERE price > 10000
  )
ORDER BY o.item, o.id;
