package d

import "database/sql"

// negative case: package function sql.Open inside a loop should NOT be flagged
func openInLoop(names []string) {
	for _, n := range names {
		_ = n
		_, _ = sql.Open("mysql", "dsn") // package function, not a method call; no diagnostic expected
	}
}
