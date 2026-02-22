package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	auto     bool
	endpoint string
	files    []string
}

func parseArgs(args []string) (config, error) {
	fs := flag.NewFlagSet("mdrefactor", flag.ContinueOnError)

	var cfg config
	fs.BoolVar(&cfg.auto, "auto", false, "Apply all changes without asking for approval")
	fs.StringVar(&cfg.endpoint, "endpoint", "http://localhost:1234", "LMStudio API endpoint")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	cfg.files = fs.Args()
	if len(cfg.files) == 0 {
		return config{}, fmt.Errorf("at least one file or directory argument is required")
	}

	return cfg, nil
}

func run() error {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	fmt.Printf("Processing %d path(s) with endpoint %s (auto=%v)\n", len(cfg.files), cfg.endpoint, cfg.auto)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
