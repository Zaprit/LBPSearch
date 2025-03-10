-- name: GetSlot :one
SELECT * FROM slot WHERE id = sqlc.arg(id);

-- name: GetSlotsSort :many
SELECT *
FROM slot
WHERE (slot."npHandle" ILIKE sqlc.arg(search_query)) OR (slot.name ILIKE sqlc.arg(search_query)) OR (slot.description ILIKE sqlc.arg(search_query))
order by (case when @search_column = 'author' and @search_direction = 'ASC' then slot."npHandle" end) asc,
         (case when @search_column = 'author' and @search_direction = 'DESC' then slot."npHandle" end) desc,
         (case when @search_column = 'hearts' and @search_direction = 'ASC' then slot."heartCount" end) asc,
         (case when @search_column = 'hearts' and @search_direction = 'DESC' then slot."heartCount" end) desc
OFFSET sqlc.arg(search_offset)
    LIMIT 50;