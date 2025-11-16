package c

import (
	"context"
	"database/sql"
)

// positive case: ExecContext on *sql.Tx inside a range loop
func positiveTx(tx *sql.Tx, ctx context.Context, items []int) error {
	for _, it := range items {
		_ = it
		if _, err := tx.ExecContext(ctx, "UPDATE t SET x=1"); err != nil { // want "potential N\\+1: database/sql method ExecContext called inside a loop"
			return err
		}
	}
	return nil
}
