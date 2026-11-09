package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	dailyDestination string
	dailySourceArgs  []string
	dailyRemoveOther bool
)

var downloadDailyCmd = &cobra.Command{
	Use:   "download-daily",
	Short: "Configure a daily systemd timer to download images of the day",
	Long: `Install a user systemd timer that runs the download once a day.

With Persistent=true, if the computer is off when the timer is due,
the download runs once on the next boot. If the computer was off for
several days, it still only runs once on boot (not once per missed day).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setupDailyTimer()
	},
}

func init() {
	rootCmd.AddCommand(downloadDailyCmd)

	downloadDailyCmd.Flags().StringVarP(
		&dailyDestination,
		"destination",
		"d",
		"",
		"Destination directory for downloaded images (default: ~/Pictures)")

	downloadDailyCmd.Flags().StringSliceVarP(
		&dailySourceArgs,
		"sources",
		"s",
		[]string{"apod", "bing", "earth-observatory", "epod", "nasa"},
		"Sources to download images from")

	downloadDailyCmd.Flags().BoolVarP(
		&dailyRemoveOther,
		"remove-other-files",
		"r",
		false,
		"Remove other files in the destination directory")
}

func setupDailyTimer() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}
	binPath := filepath.Join(binDir, "images-of-the-day")

	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}

	systemdDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	destDir := dailyDestination
	if destDir == "" {
		destDir = filepath.Join(home, "Pictures")
	}

	args := []string{"download", "-d", shellQuote(destDir)}
	for _, src := range dailySourceArgs {
		args = append(args, "-s", src)
	}
	if dailyRemoveOther {
		args = append(args, "-r")
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=Download images of the day

[Service]
Type=oneshot
ExecStart=%s %s
`, binPath, strings.Join(args, " "))

	timerContent := `[Unit]
Description=Run images-of-the-day download daily

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
`

	servicePath := filepath.Join(systemdDir, "images-of-the-day-daily.service")
	timerPath := filepath.Join(systemdDir, "images-of-the-day-daily.timer")

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service unit: %w", err)
	}
	if err := os.WriteFile(timerPath, []byte(timerContent), 0644); err != nil {
		return fmt.Errorf("failed to write timer unit: %w", err)
	}

	daemonReload := exec.Command("systemctl", "--user", "daemon-reload")
	daemonReload.Stdout = os.Stdout
	daemonReload.Stderr = os.Stderr
	if err := daemonReload.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd user manager: %w", err)
	}

	enable := exec.Command("systemctl", "--user", "enable", "--now", "images-of-the-day-daily.timer")
	enable.Stdout = os.Stdout
	enable.Stderr = os.Stderr
	if err := enable.Run(); err != nil {
		return fmt.Errorf("failed to enable and start timer: %w", err)
	}

	fmt.Printf("Daily download timer configured and enabled (runs once a day, catches up once on boot).\n")
	fmt.Printf("  Binary: %s\n", binPath)
	fmt.Printf("  Args:   %s\n", strings.Join(args, " "))
	return nil
}

// shellQuote wraps a string in single quotes, escaping any embedded single quotes,
// so it can be safely embedded in a systemd ExecStart line.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
