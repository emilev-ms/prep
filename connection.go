package prep

import (
	"context"
	"database/sql"
)

type (
	executer interface {
		Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}

	preparer interface {
		Prepare(query string) (*sql.Stmt, error)
	}

	rowQuerier interface {
		QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	}

	querier interface {
		Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}
)