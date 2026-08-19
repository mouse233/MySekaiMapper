package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/mouse233/MySekaiMapper/go/internal/mapper"
	notifyservice "github.com/mouse233/MySekaiMapper/go/internal/notify"
	httpserver "github.com/mouse233/MySekaiMapper/go/internal/server"
	"github.com/mouse233/MySekaiMapper/go/internal/service"
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
	case "serve":
		runServe(root, commandArgs)
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

func runServe(root string, args []string) {
	flags := flag.NewFlagSet("serve", flag.ExitOnError)
	host := flags.String("host", "0.0.0.0", "bind host")
	port := flags.Int("port", 9478, "bind port")
	envFile := flags.String("env", filepath.Join(root, ".env"), "dotenv file path")
	_ = flags.Parse(args)
	if *port < 1 || *port > 65535 {
		fatal(fmt.Errorf("port must be between 1 and 65535"))
	}
	if err := mapper.LoadDotEnv(*envFile); err != nil {
		fatal(err)
	}

	settings := service.SettingsFromRoot(root)
	store, err := service.NewStore(settings)
	if err != nil {
		fatal(err)
	}
	notifier := notifyservice.New(notifyservice.Config{
		BarkMapFile:       settings.BarkMapFile,
		PushMapFile:       settings.PushMapFile,
		BarkIcon:          settings.BarkIcon,
		BarkImageBase:     settings.BarkImageBase,
		FallbackImageBase: settings.FallbackImageBase,
		TelegramBotToken:  settings.TelegramBotToken,
		TelegramChatID:    settings.TelegramChatID,
	})
	processor := service.NewGenerationProcessor(settings, store, notifier)
	processor.Logf = log.Printf
	submitter := &service.AsyncSubmitter{Store: store, Processor: processor, Logf: log.Printf, Workers: 1}
	chunks, err := service.NewChunkStore(filepath.Join(settings.TmpDir, "uploads"), mapper.MaxArchiveSize)
	if err != nil {
		fatal(err)
	}
	handler, err := httpserver.New(httpserver.ConfigFromSettings(settings), chunks, submitter)
	if err != nil {
		fatal(err)
	}
	handler.Logf = log.Printf

	address := net.JoinHostPort(*host, fmt.Sprintf("%d", *port))
	server := &http.Server{
		Addr:              address,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	fmt.Printf("[READY] Go service listening on http://%s\n", address)
	fmt.Printf("[READY] POST http://%s/uploadMySekai\n", address)
	if settings.ReportEnabled {
		fmt.Printf("[READY] POST http://%s%s\n", address, settings.ReportPath)
	}
	shutdownSignal, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	shutdownDone := make(chan struct{})
	go func() {
		<-shutdownSignal.Done()
		fmt.Fprintln(os.Stderr, "[SHUTDOWN] stopping HTTP server and draining accepted jobs")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("HTTP shutdown failed: %v", err)
		}
		submitter.Close()
		close(shutdownDone)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fatal(err)
	}
	if shutdownSignal.Err() != nil {
		<-shutdownDone
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: mysekaimapper <inspect|generate|serve> [--root <repository-root>] [options]")
	fmt.Fprintln(os.Stderr, "  inspect  decrypt and print a safe aggregate summary")
	fmt.Fprintln(os.Stderr, "  generate decrypt, parse, and render map PNGs")
	fmt.Fprintln(os.Stderr, "  serve    run upload and Reqable report endpoints")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "[ERROR]", err)
	os.Exit(1)
}
