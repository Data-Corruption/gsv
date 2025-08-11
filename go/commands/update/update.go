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
		return update(ctx, version)
	},
}

func update(ctx context.Context, version string) error {

	if version == "vX.X.X" {
		fmt.Println("Dev build detected, skipping update.")
		return nil
	}

	lCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	latest, err := git.LatestGitHubReleaseTag(lCtx, RepoURL)
	if err != nil {
		return err
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
	// ensure the path is absolute
	selfPath, err := filepath.Abs(selfReal)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of executable: %w", err)
	}

	runSudo := false
	if !isRoot {
		if filepath.Dir(selfPath) == "/usr/local/bin" {
			if runSudo, err = prompt.YesNo("This update requires root privileges. Do you want to run the update with sudo?"); err != nil {
				return fmt.Errorf("failed to prompt for sudo: %w", err)
			}
			if !runSudo {
				fmt.Println("Update aborted. Please run the command with sudo to update.")
				return nil
			}
		} else {
			if filepath.Dir(selfPath) != filepath.Join(os.Getenv("HOME"), ".local", "bin") {
				if runSudo, err = prompt.YesNo("Unsure if sudo is required. Do you want to run the update with sudo?"); err != nil {
					return fmt.Errorf("failed to prompt for sudo: %w", err)
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
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}
