package main

import (
	"flag"
	"fmt"
	"github.com/KevinFasusi/sigzag/pkg/crawler"
	//t "github.com/KevinFasusi/sigzag/pkg/terminal"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := flag.String("root", ".", "Root directory")
	level := flag.Int("level", 2, "Maximum directory nesting depth")
	diffManifest := flag.Bool("diff", false, "Compare two manifests and return the difference if any")
	asset := flag.String("asset", crawler.ASSET.Strings(), "Manifests")
	history := flag.Bool("history", false, "Returns the history of an asset")
	tagFile := flag.String("tag-file", crawler.SIGZAG.Strings(), "Prepends the output file with string")
	compareManifest := flag.Bool("compare-manifest", false, "Compare manifests")
	compareMerkle := flag.Bool("compare-merkle", false, "Compare merkle")
	out := flag.String("output-file", "", "Directory for output")
	//terminal := flag.Bool("terminal", false, "Launch terminal")

	flag.Parse()
	path, err := filepath.Abs(*root)

	if *level <= 0 {
		log.Fatalf("level must be greater than 0. --level %d flag set\n", *level)
	}

	if err != nil {
		log.Fatalf("path error: %s", err)
	}
	length := len(strings.Split(path, string(os.PathSeparator)))

	levelStart := length - 1
	config := crawler.Config{
		Root:    levelStart,
		Depth:   *level + 1,
		TagFile: *tagFile,
		OutDir:  *out,
	}

	if !*compareManifest && !*compareMerkle && !*diffManifest && !*history && *asset == crawler.ASSET.Strings() {
		var m crawler.Manager
		_, _, err = m.GenerateManifest(path, config)
		if err != nil {
			fmt.Printf("error generating manifests, %s", err)
		}
	}

	if *diffManifest {
		var m crawler.Manager
		_ = m.Diff(flag.Args()[0], flag.Args()[1], false)
	}

	if *history && *asset != crawler.ASSET.Strings() {
		var m crawler.Manager
		m.History(*asset, flag.Args())
	}
	if *compareManifest {
		var m crawler.Manager
		m.Compare(flag.Args()[0], flag.Args()[1], crawler.MANIFEST)
	}
	if *compareMerkle {
		var m crawler.Manager
		m.Compare(flag.Args()[0], flag.Args()[1], crawler.MERKLETREE)
	}
}
