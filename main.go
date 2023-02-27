package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
)

// Note: this will bundle the tests as well, but since that has negligible impact on the size of the binary,
// it's preferable to filter them out from the bundle than to e.g. create a separate directory for tests
//
//go:embed policy data
var content embed.FS

func main() {
	// Remove date and time from any `log.*` calls, as that doesn't add much of value here
	log.SetFlags(0)

	// TODO: Obviously, we'll want to deal with directories and not single files, but we'll need to decide on what
	//       format to use for merging the ASTs, or if we should just present them as they are in a collection.
	if len(os.Args) < 2 {
		log.Fatal("Rego file to lint must be provided as input")
	}

	regalBundle := mustLoadRegalBundle()

	regoFile, err := loader.RegoWithOpts(os.Args[1], ast.ParserOptions{ProcessAnnotation: true})
	if err != nil {
		log.Fatal(err)
	}
	astJSON := mustJSON(regoFile.Parsed)
	var input map[string]any
	if err = json.Unmarshal(astJSON, &input); err != nil {
		log.Fatal(err)
	}

	// TODO: Make timeout configurable via command line flag
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	query, err := rego.New(
		rego.ParsedBundle("regal", &regalBundle),
		rego.Input(input),
		rego.Query("report = data.regal.main.report"),
		rego.EnablePrintStatements(true),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
	).PrepareForEval(ctx)
	if err != nil {
		log.Fatal(err)
	}

	result, err := query.Eval(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Create a reporter interface and implementations
	fmt.Println(string(mustJSON(result)))
}

func mustJSON(x any) []byte {
	bytes, err := json.Marshal(x)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

func mustLoadRegalBundle() bundle.Bundle {
	embedLoader, err := bundle.NewFSLoader(content)
	if err != nil {
		log.Fatal(err)
	}
	bundleLoader := embedLoader.WithFilter(func(abspath string, info fs.FileInfo, depth int) bool {
		return strings.HasSuffix(info.Name(), "_test.rego")
	})

	regalBundle, err := bundle.NewCustomReader(bundleLoader).
		WithSkipBundleVerification(true).
		WithProcessAnnotations(true).
		WithBundleName("regal").
		Read()

	if err != nil {
		log.Fatal(err)
	}

	return regalBundle
}
