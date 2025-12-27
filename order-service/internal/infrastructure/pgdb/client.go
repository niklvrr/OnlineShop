package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log/slog"
	"time"
)

var (
	errDbUrlEmpty = errors.New("db url is empty")

	errFailedToConnectDB = errors.New("failed to connect to db")
	errFailedToPingDB    = errors.New("failed to ping db")

	errMigrationInitFailed           = errors.New("migration run failed")
	errMigrationVersionCheckFailed   = errors.New("migration version check failed")
	errFailedToForceMigrationVersion = errors.New("failed to force migration version")
	errMigrationRunFailed            = errors.New("migration run failed")
)

func NewPgClient(dbUrl string) (*sql.DB, error) {
	if dbUrl == "" {
		return nil, errDbUrlEmpty
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConnectDB, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToPingDB, err)
	}

	return db, nil
}

func RunMigrations(dbUrl string, logger *slog.Logger) error {
	if dbUrl == "" {
		return errDbUrlEmpty
	}

	mg, err := migrate.New(
		"file://migrations",
		dbUrl,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", errMigrationInitFailed, err)
	}

	version, dirty, err := mg.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("%w: %w", errMigrationVersionCheckFailed, err)
	}

	if dirty {
		logger.Warn("database is in dirty state, forcing version", slog.Uint64("version", uint64(version)))
		if err := mg.Force(int(version)); err != nil {
			return fmt.Errorf("%w: %w", errFailedToForceMigrationVersion, err)
		}
		logger.Debug("dirty state cleared, retrying migration")
	}

	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("%w: %w", errMigrationRunFailed, err)
	}

	return nil
}
