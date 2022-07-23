package static

import (
	"os"
	"path/filepath"

	"git.sr.ht/~kota/memex/ui"
)

func Copy(outDir string) error {
	entries, err := ui.Files.ReadDir("static")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() == true {
			continue
		}

		data, err := ui.Files.ReadFile(filepath.Join("static", e.Name()))
		if err != nil {
			return err
		}

		path := filepath.Join(outDir, e.Name())
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
