package main

import (
	"fmt"
	"os"
	"path/filepath"

	"git.sr.ht/~kota/memex/convert"
	"git.sr.ht/~kota/memex/redact"
	"git.sr.ht/~kota/memex/static"
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

	inputs, err := list(inDir)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed while listing files in %s: %v\n",
			inDir,
			err,
		)
	}

	denylist, err := redact.Load(filepath.Join(inDir, "denylist.md"))
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed while loading denylist %s: %v\n",
			filepath.Join(inDir, "denylist.md"),
			err,
		)
	}
	c := convert.Converter{
		Inputs:     inputs,
		InDir:      inDir,
		OutDir:     outDir,
		Redactions: denylist,
	}
	err = c.All()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed while converting files: %v\n",
			err,
		)
	}

	err = static.Copy(outDir)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed while copying static files: %v\n",
			err,
		)
	}

	fmt.Println("memex: converted", len(inputs), "files")
}
