package cmd

import (
	"fmt"
	"images-of-the-day/sources"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	destinationDir  string
	downloadSources []sources.Source
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download images of the day from various sources",

	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}
func init() {
	destinationDir = *downloadCmd.Flags().StringP(
		"destination",
		"d",
		"~/Pictures",
		"Destination directory for downloaded images")

	sourcesArgs := downloadCmd.Flags().StringSliceP(
		"sources",
		"s",
		[]string{"bing", "nasa", "apod", "earth-observatory", "epod"},
		"Sources to download images from")
	for _, source := range *sourcesArgs {
		switch source {
		case "bing":
			downloadSources = append(downloadSources, sources.Bing{})
		case "nasa":
			downloadSources = append(downloadSources, sources.Nasa{})
		case "apod":
			downloadSources = append(downloadSources, sources.Apod{})
		case "earth-observatory":
			downloadSources = append(downloadSources, sources.EarthObservatory{})
		case "epod":
			downloadSources = append(downloadSources, sources.Epod{})
		}
	}
}

func run() {
	date := time.Now().Format("2006-01-02")

	for _, source := range downloadSources {
		links, err := source.GetImageLinks()
		if err != nil {
			println(err.Error())
		}
		for i, link := range links {
			response, err := http.Get(link)
			if err != nil {
				println(err.Error())
				continue
			}
			defer response.Body.Close()

			fileName := fmt.Sprintf("%s/%s_%s_%d.jpg", destinationDir, date, source.GetName(), i)
			f, err := os.Create(fileName)
			if err != nil {
				println(err.Error())
				continue
			}
			defer f.Close()

			_, err = io.Copy(f, response.Body)
			if err != nil {
				println(err.Error())
				continue
			}
			println("Downloaded", fileName)
		}
	}
}
