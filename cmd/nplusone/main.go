package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"govet-nplusone/internal/analyzer/nplusone"
)

// main runs the analyzer as a standalone vet-style tool.
func main() {
	singlechecker.Main(nplusone.Analyzer)
}
