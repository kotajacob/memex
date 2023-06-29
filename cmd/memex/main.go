package main

import (
	"fmt"
	"os"
	"path/filepath"

	"git.sr.ht/~kota/memex/assets"
	"git.sr.ht/~kota/memex/convert"
	"git.sr.ht/~kota/memex/redact"
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: memex [memex_path] [output_path]")
	os.Exit(1)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	inDir, outDir := os.Args[1], os.Args[2]

	am, err := assets.Copy(outDir)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed while copying static files: %v\n",
			err,
		)
	}
	favicon := am["favicon.png"]
	css := am["main.css"]

	files, err := convert.List(inDir)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed while listing files in %s: %v\n",
			inDir,
			err,
		)
	}

	denylist, err := redact.Load(filepath.Join(inDir, "redactions.md"))
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed while loading denylist %s: %v\n",
			filepath.Join(inDir, "denylist.md"),
			err,
		)
	}
	c := convert.Converter{
		InDir:      inDir,
		OutDir:     outDir,
		Redactions: denylist,
		Favicon:    favicon,
		CSS:        css,
	}
	err = c.All(files)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed while converting files: %v\n",
			err,
		)
	}
	fmt.Println("memex: converted", len(files), "files")
}
