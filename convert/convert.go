package convert

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"git.sr.ht/~kota/memex/links"
	"git.sr.ht/~kota/memex/normalize"
	"git.sr.ht/~kota/memex/redact"
	"git.sr.ht/~kota/memex/ui"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

const mainTMPL = "main.tmpl"

// Page holds the information needed to render an individual page.
type Page struct {
	// Filename is the name without an extension or full path.
	Filename string

	// Content is the HTML content to place in the page.
	Content string

	// From is a list of other pages that link to this page.
	From map[string]struct{}

	// Favicon is an href for the site's favicon.
	// It's name contains a hash to allow for immutable caching.
	Favicon string

	// CSS is an href for the site's CSS.
	// It's name contains a hash to allow for immutable caching.
	CSS string
}

// File holds information about a file on disk.
type File struct {
	Path    string
	ModTime time.Time
}

// Converter holds all the information needed to convert a list of input files.
type Converter struct {
	InDir  string
	OutDir string

	// Redactions is a list on substrings which, if matched with a filename,
	// prevent the file from being exported.
	Redactions []string

	// Favicon is an href for the site's favicon.
	// It's name contains a hash to allow for immutable caching.
	Favicon string

	// CSS is an href for the site's CSS.
	// It's name contains a hash to allow for immutable caching.
	CSS string
}

// All processes a list of input Files, cleaning up and converting each one to
// html and writing the final output into the outDir. The inDir portion of
// each filepath is replaced with outDir; so subdirectory structures remain
// unchanged.
func (c Converter) All(files []File) error {
	var mediaFiles []File
	var markdownFiles []File
	linkMap := make(map[string]map[string]struct{})
	for _, f := range files {
		if !strings.HasSuffix(f.Path, ".md") {
			mediaFiles = append(mediaFiles, f)
			continue
		}

		if redact.Hidden(
			strings.TrimPrefix(f.Path, c.InDir+"/"),
			c.Redactions,
		) {
			continue
		}

		markdownFiles = append(markdownFiles, f)
		linkMap[toName(f.Path, c.InDir)] = make(map[string]struct{})
	}

	for _, f := range markdownFiles {
		data, err := os.ReadFile(f.Path)
		if err != nil {
			return err
		}

		name := toName(f.Path, c.InDir)
		linkMap = links.Map(data, name, linkMap)
	}

	for _, f := range mediaFiles {
		if err := c.media(f); err != nil {
			return err
		}
	}

	for _, f := range markdownFiles {
		if err := c.markdown(f.Path, linkMap); err != nil {
			return err
		}
	}
	return nil
}

// markdown handles the conversion and writing of a markdown file.
func (c Converter) markdown(
	input string,
	linkMap map[string]map[string]struct{},
) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}

	// Make a few modifications.
	data = redact.Redact(input, data, c.Redactions)
	data = links.Modify(data, linkMap)

	// Convert to html.
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(data, &buf); err != nil {
		return fmt.Errorf("failed to convert markdown to html: %v", err)
	}

	// Apply template.
	name := toName(input, c.InDir)
	page := Page{
		Filename: name,
		Content:  buf.String(),
		From:     linkMap[name],
		Favicon:  c.Favicon,
		CSS:      c.CSS,
	}

	tmpl, err := template.New(mainTMPL).
		Funcs(template.FuncMap{"Normalize": normalize.String}).
		ParseFS(ui.Files, mainTMPL)
	if err != nil {
		return fmt.Errorf("failed to load main.tmpl: %v", err)
	}

	buf.Reset()
	err = tmpl.Execute(&buf, page)
	if err != nil {
		return fmt.Errorf(
			"failed while executing template on %s: %v",
			input,
			err,
		)
	}
	data = buf.Bytes()

	// Write file.
	path := filepath.Join(c.OutDir, toName(input, c.InDir))
	path = normalize.String(path) + ".html"
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// media copies a media file from inDir to the outDir.
func (c Converter) media(f File) error {
	outPath := strings.TrimPrefix(f.Path, c.InDir)
	outPath = filepath.Join(c.OutDir, outPath)

	// Skip if the output file is newer than the input file.
	if info, err := os.Stat(outPath); err == nil {
		if info.ModTime().After(f.ModTime) {
			return nil
		}
	}

	data, err := os.ReadFile(f.Path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, data, 0o644)
}

func toName(path, inDir string) string {
	path = strings.TrimSuffix(path, ".md")
	return strings.TrimPrefix(path, inDir+"/")
}
