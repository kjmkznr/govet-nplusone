package a

import (
	"context"
	"database/sql"
)

// positive case: method call from database/sql inside a for loop
func positive(db *sql.DB, ctx context.Context, ids []int) {
	for _, id := range ids {
		_ = id
		_ = db.QueryRowContext(ctx, "SELECT 1") // want "potential N\\+1: database/sql method QueryRowContext called inside a loop"
	}
}
