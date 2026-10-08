package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func getMigrationPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(1)
	dir := filepath.Dir(currentFile)
	// internal/db -> internal -> apps/backend
	backendRoot := filepath.Dir(filepath.Dir(dir))
	return filepath.Join(backendRoot, "migrations", filename)
}

func runMigration(database *sql.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = database.ExecContext(ctx, string(content))
	if err != nil {
		return fmt.Errorf("execute migration %s: %w", path, err)
	}

	return nil
}

func InitDB(database *sql.DB) error {
	if database == nil {
		return fmt.Errorf("database is nil")
	}

	migrations := []string{
		"000_create_enums.sql",
		"001_create_users.sql",
		"002_create_projects.sql",
		"003_create_project_members.sql",
		"004_create_project_join_requests.sql",
		"005_create_incidents.sql",
		"006_create_indexes.sql",
	}

	for _, migration := range migrations {
		if err := runMigration(database, getMigrationPath(migration)); err != nil {
			return err
		}
	}

	DB = database

	return nil
}

func Connect() (*sql.DB, error) {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		return nil, err
	}

	// Configure a sensible connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify DB is reachable with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
