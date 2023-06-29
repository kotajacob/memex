package convert

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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

// Converter holds all the information needed to convert a list of input files.
type Converter struct {
	Inputs []string
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

// All processes a list of input files, cleaning up and converting each one to
// html and writing the final output into the outDir. The inDir portion of
// each filepath is replaced with outDir; so subdirectory structures remain
// unchanged.
func (c Converter) All() error {
	var mediaFiles []string
	var markdownFiles []string
	linkMap := make(map[string]map[string]struct{})
	for _, i := range c.Inputs {
		if !strings.HasSuffix(i, ".md") {
			mediaFiles = append(mediaFiles, i)
			continue
		}

		if redact.Hidden(
			strings.TrimPrefix(i, c.InDir+"/"),
			c.Redactions,
		) {
			continue
		}

		markdownFiles = append(markdownFiles, i)
		linkMap[toName(i, c.InDir)] = make(map[string]struct{})
	}

	for _, i := range markdownFiles {
		data, err := os.ReadFile(i)
		if err != nil {
			return err
		}

		name := toName(i, c.InDir)
		linkMap = links.Map(data, name, linkMap)
	}

	for _, i := range mediaFiles {
		if err := c.media(i); err != nil {
			return err
		}
	}

	for _, i := range markdownFiles {
		if err := c.markdown(i, linkMap); err != nil {
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
func (c Converter) media(input string) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}

	path := strings.TrimPrefix(input, c.InDir)
	path = filepath.Join(c.OutDir, path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func toName(path, inDir string) string {
	path = strings.TrimSuffix(path, ".md")
	return strings.TrimPrefix(path, inDir+"/")
}
