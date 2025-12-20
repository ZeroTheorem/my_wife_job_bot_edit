-- name: CreateRow :exec
INSERT INTO notes (name, val, month, year)
VALUES (?, ?, ?, ?);

-- name: GetAvg :one
SELECT AVG(val) FROM notes
WHERE name = ? AND month = ? AND year = ?;

-- name: DeleteLastRow :one
DELETE FROM notes
WHERE id = (SELECT MAX(id) from notes)
RETURNING name, val;

-- name: GetWifeSalary :one
SELECT COUNT(*), SUM(val)
FROM notes
WHERE name = ? AND month = ? AND year = ?;

-- name: GetMonthlyTotal :one
SELECT SUM(val)
FROM notes
WHERE month = ? AND year = ?;

-- name: GetAllRowsInMonth :many
SELECT *
FROM notes
WHERE month = ? AND year = ?;
