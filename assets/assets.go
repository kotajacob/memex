package assets

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.sr.ht/~kota/memex/ui"
)

// Copy the static files to outDir, renaming them to place a sha1 hash between
// their basename and extension. A map of original names to hash-included names
// is returned.
func Copy(outDir string) (map[string]string, error) {
	entries, err := ui.Files.ReadDir("assets")
	if err != nil {
		return nil, err
	}

	hashedNames := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() == true {
			continue
		}

		data, err := ui.Files.ReadFile(filepath.Join("assets", e.Name()))
		if err != nil {
			return nil, err
		}

		h := sha1.New()
		h.Write(data)
		sum := fmt.Sprintf("%x", h.Sum(nil))

		name := e.Name()
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		hashedName := base + "." + sum + ext
		hashedNames[name] = hashedName

		// ext already contains a . prefix.
		path := filepath.Join(outDir, hashedName)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return nil, err
		}

		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, err
		}
	}
	return hashedNames, nil
}
