// Package golden provides shared golden-fixture loading helpers used by test
// packages across the tree (DNP3-097). It is a leaf package with no internal
// dependencies, so it can be imported by low-level packages (e.g. the link
// frame package) without creating an import cycle.
package golden

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Dir returns the absolute path to the shared golden-fixture directory
// (active_work/testdata) relative to the repository root, regardless of which
// package calls it. It is the single source of truth for golden fixture
// locations so test packages do not each reimplement the path/decode logic.
func Dir() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed")
	}
	// internal/testutils/golden/golden.go → repo root is three parents up.
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(thisFile))))
	return filepath.Abs(filepath.Join(root, "active_work", "testdata"))
}

// LoadHex reads a .hex fixture from active_work/testdata, strips '#' line
// comments and all whitespace, and hex-decodes the remaining bytes. It is the
// shared golden loader used by both the master and link-frame test packages
// (DNP3-097: no duplicated golden-loader logic).
func LoadHex(name string) ([]byte, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	var clean strings.Builder
	for _, line := range strings.Split(string(raw), "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		clean.WriteString(line)
	}
	s := strings.ReplaceAll(clean.String(), " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\r", "")
	return hex.DecodeString(s)
}
