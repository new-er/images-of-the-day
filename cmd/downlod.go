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
	destinationDir string
	sourceArgs     []string
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download images of the day from various sources",

	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	downloadCmd.Flags().StringVarP(
		&destinationDir,
		"destination",
		"d",
		"~/Pictures",
		"Destination directory for downloaded images")

	downloadCmd.Flags().StringSliceVarP(
		&sourceArgs,
		"sources",
		"s",
		[]string{"bing", "nasa", "apod", "earth-observatory", "epod"},
		"Sources to download images from")
}

func run() {
	s := []sources.Source{}
	for _, source := range sourceArgs {
		switch source {
		case "bing":
			s = append(s, sources.Bing{})
		case "nasa":
			s = append(s, sources.Nasa{})
		case "apod":
			s = append(s, sources.Apod{})
		case "earth-observatory":
			s = append(s, sources.EarthObservatory{})
		case "epod":
			s = append(s, sources.Epod{})
		}
	}

	date := time.Now().Format("2006-01-02")
	if _, err := os.Stat(destinationDir); os.IsNotExist(err) {
		os.MkdirAll(destinationDir, os.ModePerm)
	}

	for _, source := range s {
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
