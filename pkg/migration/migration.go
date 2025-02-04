package migration

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(db *sql.DB) error {

	migrationDir := "./migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return err
	}

	var migrations []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "_up.sql") {
			migrations = append(migrations, file.Name())
		}
	}

	sort.Strings(migrations)

	for _, migration := range migrations {
		path := filepath.Join(migrationDir, migration)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		if _, err = tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing migration %s: %v", migration, err)
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		log.Printf("Applied migration: %s\n", migration)
	}

	return nil
}
