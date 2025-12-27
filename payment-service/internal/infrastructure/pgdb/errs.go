package pgdb

import (
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func handleDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrDuplicate
		case "23503":
			return ErrForeignKeyViolation
		}
	}

	return err
}

var (
	ErrNotFound            = errors.New("record not found")
	ErrDuplicate            = errors.New("duplicate record")
	ErrForeignKeyViolation  = errors.New("foreign key violation")
)

