package b

import (
	"context"
	"database/sql"
)

// negative case: database/sql call but not inside a loop -> no diagnostics expected
func negativeOutsideLoop(db *sql.DB, ctx context.Context) {
	_ = db.QueryRowContext(ctx, "SELECT 1")
}
