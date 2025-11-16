package nplusone

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNPlusOne(t *testing.T) {
	testdata := analysistest.TestData()
	// Run against all test packages under testdata/src
	analysistest.Run(t, testdata, Analyzer, "a", "b", "c", "d")
}

// Guard to ensure the path is correct during local runs as well.
// Some editors may run tests from repo root; the above uses analysistest.TestData()
// which finds the nearest "testdata" directory relative to this test file.
// This dummy reference keeps linters quiet about unused imports in some setups.
var _ = filepath.Separator
