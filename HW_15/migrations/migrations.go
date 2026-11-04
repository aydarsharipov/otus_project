package migrations

import "embed"

// sql файлы миграций вшиваем прямо в программу
//
//go:embed *.sql
var FS embed.FS
