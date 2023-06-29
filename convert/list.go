package convert

import (
	"os"
	"path"
	"strings"
)

// List recursively finds all non-hidden and non-empty Files in a directory
// returning either a full list or an error.
func List(name string) ([]File, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}

	var files []File
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
			subPaths, err := List(path.Join(name, e.Name()))
			if err != nil {
				return nil, err
			}
			files = append(files, subPaths...)
		} else {
			files = append(
				files,
				File{
					Path:    path.Join(name, e.Name()),
					ModTime: info.ModTime(),
				},
			)
		}
	}
	return files, nil
}
