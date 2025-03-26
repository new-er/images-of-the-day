package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "images-of-the-day",
	Short: "Download images of the day from various sources",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize()
	rootCmd.AddCommand(downloadCmd)

}
