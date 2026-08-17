package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mouse233/MySekaiMapper/go/internal/mapper"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	root, commandArgs, err := resolveRoot(cwd, os.Args[2:])
	if err != nil {
		fatal(err)
	}

	switch os.Args[1] {
	case "inspect":
		runInspect(root, commandArgs)
	case "generate":
		runGenerate(root, commandArgs)
	default:
		usage()
		os.Exit(2)
	}
}

// resolveRoot accepts --root anywhere after the subcommand so a compiled
// binary can run outside a source checkout. Other flags remain for the
// command-specific FlagSet to parse.
func resolveRoot(cwd string, args []string) (string, []string, error) {
	filtered := make([]string, 0, len(args))
	rootValue := ""
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch {
		case argument == "--root":
			if index+1 >= len(args) {
				return "", nil, fmt.Errorf("--root requires a directory")
			}
			rootValue = args[index+1]
			index++
		case strings.HasPrefix(argument, "--root="):
			rootValue = strings.TrimPrefix(argument, "--root=")
			if rootValue == "" {
				return "", nil, fmt.Errorf("--root requires a directory")
			}
		default:
			filtered = append(filtered, argument)
		}
	}
	if rootValue == "" {
		root, err := mapper.DiscoverRoot(cwd)
		return root, filtered, err
	}
	root, err := filepath.Abs(rootValue)
	if err != nil {
		return "", nil, fmt.Errorf("resolve --root: %w", err)
	}
	if _, err := os.Stat(filepath.Join(root, "assets", "resourceId.csv")); err != nil {
		return "", nil, fmt.Errorf("invalid --root %q: missing assets/resourceId.csv", root)
	}
	return root, filtered, nil
}

func runInspect(root string, args []string) {
	flags := flag.NewFlagSet("inspect", flag.ExitOnError)
	input := flags.String("input", "", "encrypted .bin save path (required)")
	envFile := flags.String("env", filepath.Join(root, ".env"), "dotenv file path")
	_ = flags.Parse(args)
	if *input == "" {
		flags.Usage()
		os.Exit(2)
	}
	if err := mapper.LoadDotEnv(*envFile); err != nil {
		fatal(err)
	}
	drops, err := mapper.ReadDrops(*input)
	if err != nil {
		fatal(err)
	}
	result, err := json.Marshal(mapper.Summarize(drops))
	if err != nil {
		fatal(err)
	}
	fmt.Println(string(result))
}

func runGenerate(root string, args []string) {
	flags := flag.NewFlagSet("generate", flag.ExitOnError)
	input := flags.String("input", "", "encrypted .bin save path (required)")
	output := flags.String("output", filepath.Join(root, "data", "go-rewrite-output", "latest"), "output directory")
	assets := flags.String("assets", filepath.Join(root, "assets"), "asset directory")
	envFile := flags.String("env", filepath.Join(root, ".env"), "dotenv file path")
	_ = flags.Parse(args)
	if *input == "" {
		flags.Usage()
		os.Exit(2)
	}
	if err := mapper.LoadDotEnv(*envFile); err != nil {
		fatal(err)
	}

	drops, err := mapper.ReadDrops(*input)
	if err != nil {
		fatal(err)
	}
	result, err := mapper.Generate(drops, filepath.Join(*assets, "resourceId.csv"), filepath.Join(*assets, "NotoSansSC-Regular.ttf"), *output)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("[DONE] Generated %d maps in %s (%d drops)\n", len(result.MapFiles), *output, result.Summary.TotalDrops)
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: mysekaimapper <inspect|generate> [--root <repository-root>] [options]")
	fmt.Fprintln(os.Stderr, "  inspect  decrypt and print a safe aggregate summary")
	fmt.Fprintln(os.Stderr, "  generate decrypt, parse, and render map PNGs")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "[ERROR]", err)
	os.Exit(1)
}
