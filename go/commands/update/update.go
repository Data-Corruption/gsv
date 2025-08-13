package update

import (
	"context"
	"fmt"
	"gsv/go/system/git"
	"gsv/go/x"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Data-Corruption/stdx/xlog"
	"github.com/Data-Corruption/stdx/xterm/prompt"
	"github.com/urfave/cli/v3"
	"golang.org/x/mod/semver"
)

// Template variables ---------------------------------------------------------

const (
	RepoURL          = "https://github.com/Data-Corruption/gsv.git"
	InstallScriptURL = "https://raw.githubusercontent.com/Data-Corruption/gsv/main/scripts/install.sh"
)

// ----------------------------------------------------------------------------

var Command = &cli.Command{
	Name:  "update",
	Usage: "update to the latest release",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		// get current version from context
		version, ok := ctx.Value("appVersion").(string)
		if !ok {
			return fmt.Errorf("failed to get appVersion from context")
		}
		// update
		return update(ctx, version, cmd.Bool("yes"))
	},
}

func update(ctx context.Context, version string, autoYes bool) error {

	if version == "vX.X.X" {
		fmt.Println("Dev build detected, skipping update.")
		return nil
	}

	lCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	latest, err := git.LatestGitHubReleaseTag(lCtx, RepoURL)
	if err != nil {
		// if the context timed out, surface a clearer message but do not block other commands
		if lCtx.Err() != nil {
			return fmt.Errorf("failed to check latest release: %w", lCtx.Err())
		}
		return fmt.Errorf("failed to check latest release: %w", err)
	}

	updateAvailable := semver.Compare(latest, version) > 0
	if !updateAvailable {
		fmt.Println("No updates available.")
		return nil
	}
	xlog.Infof(ctx, "Found new version: %s", latest)
	fmt.Printf("Found new version: %s\n", latest)

	// get if sudo
	isRoot := os.Geteuid() == 0

	// get the executable path
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	selfReal, errSelf := filepath.EvalSymlinks(self)
	if errSelf != nil {
		selfReal = self // fallback to self if symlink resolution fails
	}
	selfPath, err := filepath.Abs(selfReal) // ensure the path is absolute
	if err != nil {
		return fmt.Errorf("failed to get absolute path of executable: %w", err)
	}

	// check if sudo is required
	runSudo := false
	if !isRoot {
		homeDir, herr := os.UserHomeDir()
		if herr != nil || homeDir == "" {
			homeDir = os.Getenv("HOME")
		}
		if filepath.Dir(selfPath) == "/usr/local/bin" {
			if autoYes {
				runSudo = true
			} else {
				if runSudo, err = prompt.YesNo("This update requires root privileges. Do you want to run the update with sudo?"); err != nil {
					return fmt.Errorf("failed to prompt for sudo: %w", err)
				}
				if !runSudo {
					fmt.Println("Update aborted. Please run the command with sudo to update.")
					return nil
				}
			}
		} else {
			if filepath.Dir(selfPath) != filepath.Join(homeDir, ".local", "bin") {
				if autoYes {
					runSudo = false // not sure which is better, going with this for now
				} else {
					if runSudo, err = prompt.YesNo("Unsure if sudo is required. Do you want to run the update with sudo?"); err != nil {
						return fmt.Errorf("failed to prompt for sudo: %w", err)
					}
				}
			}
		}
	}

	// run the install command
	pipeline := fmt.Sprintf("curl -sSfL %s | %sbash -s -- latest %q", InstallScriptURL, x.Ternary(runSudo, "sudo ", ""), filepath.Dir(selfPath))
	xlog.Debugf(ctx, "Running update command: %s", pipeline)

	iCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(iCtx, "bash", "-c", pipeline)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}
