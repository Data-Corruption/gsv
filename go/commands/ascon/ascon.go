package ascon

import (
	"context"
	"fmt"
	"os"
	"sprout/go/x"

	"github.com/Data-Corruption/stdx/xlog"
	"github.com/Data-Corruption/stdx/xterm/prompt"
	"github.com/urfave/cli/v3"
)

var Ascon = &cli.Command{
	Name:      "ascon",
	Usage:     "generate a SystemVerilog ROM from an Ascon KAT file",
	UsageText: "ascon <in-path> [out-path] [--sb]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "swap-bytes",
			Aliases: []string{"sb"},
			Value:   false,
			Usage:   "swap the byte order of the data portion (padding unchanged)",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		// ---- Parse args ----

		var inPath string
		var outPath string
		var swap ByteSwapMode

		args := cmd.Args()
		if args.Len() < 1 {
			fmt.Println("missing required <in-path>")
			return fmt.Errorf("missing required <in-path>")
		}
		inPath = args.Get(0)

		if args.Len() > 1 {
			outPath = args.Get(1)
		} else {
			outPath = "ascon_rom.sv"
		}

		// if outPath exists, ask for confirmation
		if _, err := os.Stat(outPath); err == nil {
			if !cmd.Bool("yes") {
				answer, err := prompt.YesNo(fmt.Sprintf("File %s already exists. Overwrite?", outPath))
				if err != nil {
					return fmt.Errorf("failed to prompt for overwrite: %w", err)
				}
				if !answer {
					return fmt.Errorf("file %s already exists", outPath)
				}
			}
		}

		swap = x.Ternary(cmd.Bool("swap-bytes"), SwapDataReverse, SwapNone)

		xlog.Infof(ctx, "Generating Ascon ROM from %s to %s with swap-bytes: %t", inPath, outPath, cmd.Bool("swap-bytes"))

		// ---- Generate rom ----

		// parse input
		vectors, err := parse(ctx, inPath)
		if err != nil {
			return fmt.Errorf("failed to parse KAT file: %w", err)
		}
		xlog.Infof(ctx, "parsed %d vectors", len(vectors))

		// generate sv code
		code := generate(ctx, vectors, swap)
		xlog.Infof(ctx, "generated %d bytes of code", len(code))

		// write output
		tmpPath := outPath + ".tmp"
		if err := os.WriteFile(tmpPath, []byte(code), 0o644); err != nil {
			return fmt.Errorf("failed to write temporary file: %w", err)
		}
		if err := os.Rename(tmpPath, outPath); err != nil {
			return fmt.Errorf("failed to rename temporary file: %w", err)
		}
		xlog.Infof(ctx, "wrote %s", outPath)

		return nil
	},
}
