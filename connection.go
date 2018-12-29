package prep

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type (
	Row struct {
		row *sqlx.Row
		err error
	}

	QueryContext interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	}

	QueryRowContext interface {
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *Row
	}

	PrepareContext interface {
		PrepareContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	}

	PrepareNamedContext interface {
		PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	}

	ExecContext interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}

	GetContext interface {
		GetContext(ctx context.Context, destination interface{}, query string, args ...interface{}) error
	}

	SelectContext interface {
		SelectContext(ctx context.Context, destination interface{}, query string, args ...interface{}) error
	}

	NamedExecContext interface {
		NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	}

	NamedQueryContext interface {
		NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	}
)