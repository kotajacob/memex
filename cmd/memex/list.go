package main

import (
	"os"
	"path"
	"strings"
)

// list recursively finds all non-hidden files in a directory returning either a
// full list or an error.
func list(name string) ([]string, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		// Ignore hidden.
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		// Ignore empty files.
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		if info.Size() == 0 {
			continue
		}

		if e.IsDir() {
			subPaths, err := list(path.Join(name, e.Name()))
			if err != nil {
				return nil, err
			}
			paths = append(paths, subPaths...)
		} else {
			paths = append(paths, path.Join(name, e.Name()))
		}
	}
	return paths, nil
}
