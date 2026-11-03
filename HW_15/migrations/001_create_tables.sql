-- +goose Up
CREATE TABLE players (
    id    INTEGER PRIMARY KEY,
    name  TEXT NOT NULL,
    color TEXT NOT NULL
);

CREATE TABLE games (
    id       INTEGER PRIMARY KEY,
    white_id INTEGER NOT NULL REFERENCES players (id),
    black_id INTEGER NOT NULL REFERENCES players (id),
    board    JSONB NOT NULL, -- доску храним целиком как json
    turn     TEXT NOT NULL,
    status   TEXT NOT NULL,
    winner   TEXT NOT NULL DEFAULT ''
);

-- при удалении игры её ходы удаляются тоже
CREATE TABLE moves (
    id       INTEGER PRIMARY KEY,
    game_id  INTEGER NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    from_row INTEGER NOT NULL,
    from_col INTEGER NOT NULL,
    to_row   INTEGER NOT NULL,
    to_col   INTEGER NOT NULL,
    piece    JSONB NOT NULL,
    notation TEXT NOT NULL
);

-- +goose Down
DROP TABLE moves;
DROP TABLE games;
DROP TABLE players;
