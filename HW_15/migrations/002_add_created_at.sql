-- +goose Up
-- новые поля: когда создали игрока и игру
ALTER TABLE players ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT now();
ALTER TABLE games ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE games DROP COLUMN created_at;
ALTER TABLE players DROP COLUMN created_at;
