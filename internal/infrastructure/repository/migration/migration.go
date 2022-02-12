package migration

import (
	"database/sql"
	"embed"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/pressly/goose/v3"
)

//go:embed data/*.sql
var embedMigrations embed.FS

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetTableName("public.goose_db_version")

	if err := goose.Up(db, "data"); err != nil {
		return err
	}
	return nil
}
