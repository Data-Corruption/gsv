package main

import (
	"context"
	"fmt"
	"gsv/go/commands/ascon"
	"gsv/go/commands/update"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Data-Corruption/stdx/xlog"
	"github.com/urfave/cli/v3"
)

// Template variables ---------------------------------------------------------

const Name = "gsv"

// ----------------------------------------------------------------------------

const DefaultLogLevel = "warn"

var Version string // set by build script

func main() {
	// Init context

	// base context with interrupt/termination handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// insert version under "appVersion" for update command
	ctx = context.WithValue(ctx, "appVersion", Version)

	// get log path (home/.<app_name>/logs)
	homeDir, herr := os.UserHomeDir()
	if herr != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}
	if homeDir == "" {
		homeDir = "." // fallback to current directory
	}
	logPath := filepath.Join(homeDir, "."+Name, "logs")
	if err := os.MkdirAll(logPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log path: %s\n", err)
		os.Exit(1)
	}

	// init logger
	log, err := xlog.New(filepath.Join(logPath, "logs"), DefaultLogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %s\n", err)
		os.Exit(1)
	}
	ctx = xlog.IntoContext(ctx, log)
	defer log.Close()

	// Init app

	app := &cli.Command{
		Name:    Name,
		Version: Version,
		Usage:   "collection of miscellaneous SystemVerilog code generators",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log",
				Value: DefaultLogLevel,
				Usage: "set log level (debug|info|warn|error|none)",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "answer yes to all prompts",
			},
		},
		Commands: []*cli.Command{
			ascon.Command,
			update.Command,
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			logLevel := cmd.String("log")
			if logLevel != DefaultLogLevel {
				if err := log.SetLevel(logLevel); err != nil {
					return ctx, err
				}
			}
			return ctx, nil
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		log.Error(err)
		fmt.Fprintln(os.Stderr, err)
		defer os.Exit(1) // defer to ensure log / other cleanup is done
	}
}
