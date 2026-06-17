package database

import (
	"context"
	"database/sql"
)

// DBTX is the interface satisfied by both *sql.DB and *sql.Tx,
// allowing all queries to run inside or outside a transaction.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// Queries holds a DBTX reference and exposes all named SQL query methods.
type Queries struct {
	db DBTX
}

// New returns a Queries backed by the provided DBTX.
// Pass database.DB (the *sql.DB pool) for normal use, or a *sql.Tx for transactions.
func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// WithTx returns a new Queries scoped to the given transaction.
func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{db: tx}
}
