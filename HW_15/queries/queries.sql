-- запросы для sqlc, по ним генерируется код в internal/db

-- игроки

-- name: NextPlayerID :one
SELECT (COALESCE(MAX(id), 0) + 1)::int AS next_id FROM players;

-- name: CreatePlayer :exec
INSERT INTO players (id, name, color) VALUES ($1, $2, $3);

-- name: GetPlayer :one
SELECT * FROM players WHERE id = $1;

-- name: ListPlayers :many
SELECT * FROM players ORDER BY id;

-- name: UpdatePlayer :exec
UPDATE players SET name = $2, color = $3 WHERE id = $1;

-- name: DeletePlayer :execrows
DELETE FROM players WHERE id = $1;

-- игры

-- name: NextGameID :one
SELECT (COALESCE(MAX(id), 0) + 1)::int AS next_id FROM games;

-- name: CreateGame :exec
INSERT INTO games (id, white_id, black_id, board, turn, status, winner)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetGame :one
SELECT * FROM games WHERE id = $1;

-- name: ListGames :many
SELECT * FROM games ORDER BY id;

-- name: UpdateGame :exec
UPDATE games SET board = $2, turn = $3, status = $4, winner = $5 WHERE id = $1;

-- name: DeleteGame :execrows
DELETE FROM games WHERE id = $1;

-- ходы

-- name: NextMoveID :one
SELECT (COALESCE(MAX(id), 0) + 1)::int AS next_id FROM moves;

-- name: CreateMove :exec
INSERT INTO moves (id, game_id, from_row, from_col, to_row, to_col, piece, notation)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetMove :one
SELECT * FROM moves WHERE id = $1;

-- name: ListMoves :many
SELECT * FROM moves ORDER BY id;

-- name: ListGameMoves :many
SELECT * FROM moves WHERE game_id = $1 ORDER BY id;
