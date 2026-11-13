package cmd

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var dailyLogsLines int

var downloadDailyLogsCmd = &cobra.Command{
	Use:   "download-daily-logs",
	Short: "Show the logs of the last runs of the daily download service",
	Long: `Show recent journal logs for the images-of-the-day-daily service
installed by the download-daily command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		journalctl := exec.Command("journalctl", "--user",
			"-u", "images-of-the-day-daily.service",
			"-n", strconv.Itoa(dailyLogsLines),
			"--no-pager", "-o", "short-iso")
		// COLUMNS large so long log lines are not truncated to terminal width.
		journalctl.Env = append(os.Environ(), "COLUMNS=999")
		journalctl.Stdout = os.Stdout
		journalctl.Stderr = os.Stderr
		if err := journalctl.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadDailyLogsCmd)

	downloadDailyLogsCmd.Flags().IntVarP(
		&dailyLogsLines,
		"lines",
		"n",
		50,
		"Number of log lines to show (default: 50)")
}
