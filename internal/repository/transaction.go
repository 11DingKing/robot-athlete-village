package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Transaction struct{ tx *sql.Tx }

func Begin(ctx context.Context, db *sql.DB) (*Transaction, error) {
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	return &Transaction{tx: tx}, nil
}
func (t *Transaction) Exec(ctx context.Context, query string, args ...any) error {
	if t == nil || t.tx == nil {
		return fmt.Errorf("transaction unavailable")
	}
	_, e := t.tx.ExecContext(ctx, query, args...)
	return e
}
func (t *Transaction) Commit() error {
	if t == nil || t.tx == nil {
		return fmt.Errorf("transaction unavailable")
	}
	return t.tx.Commit()
}
func (t *Transaction) Rollback() error {
	if t == nil || t.tx == nil {
		return nil
	}
	return t.tx.Rollback()
}
func (t *Transaction) RecordMarker(ctx context.Context, requestID string) error {
	return t.Exec(ctx, "INSERT INTO audit_events(entity_type,entity_id,action,result,request_id,created_at) VALUES('transaction','0','marker','success',?,?)", requestID, time.Now().UTC().Format(time.RFC3339))
}
