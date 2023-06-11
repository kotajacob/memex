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
}

// Converter holds all the information needed to convert a list of input files.
type Converter struct {
	Inputs     []string
	InDir      string
	OutDir     string
	Redactions []string
}

// All processes a list of input files, cleaning up and converting each one to
// html and writing the final output into the outDir. The inDir portion of
// each filepath is replaced with outDir; so subdirectory structures remain
// unchanged.
func (c Converter) All() error {
	// Filter a list of markdown files.
	var markdowns []string

	// We need to create a linkMap which maps each page to a list of other
	// pages which link to it. First we can simply populate the map with all
	// the names of the markdown pages.
	linkMap := make(map[string]map[string]struct{})
	for _, input := range c.Inputs {
		if !strings.HasSuffix(input, ".md") {
			continue
		}

		if redact.Hidden(
			strings.TrimPrefix(input, c.InDir+"/"),
			c.Redactions,
		) {
			continue
		}

		markdowns = append(markdowns, input)
		linkMap[toName(input, c.InDir)] = make(map[string]struct{})
	}

	// Populate the linkMap by writing a list of other pages that link to each
	// page.
	for _, input := range markdowns {

		data, err := os.ReadFile(input)
		if err != nil {
			return err
		}

		name := toName(input, c.InDir)
		linkMap = links.Map(data, name, linkMap)
	}

	for _, input := range c.Inputs {
		err := c.convert(input, linkMap)
		if err != nil {
			return err
		}
	}
	return nil
}

// convert will do any needed processing to a source file and then write it to
// the outDir.
func (c Converter) convert(
	input string,
	linkMap map[string]map[string]struct{},
) error {
	if !strings.HasSuffix(input, ".md") {
		if err := c.media(input); err != nil {
			return err
		}
		return nil
	}
	return c.markdown(input, linkMap)
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
