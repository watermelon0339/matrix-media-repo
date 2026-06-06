package test_internals

import (
	"path/filepath"
	"runtime"
)

func repoRootDir() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to determine repository root")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}
