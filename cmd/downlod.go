package cmd

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/new-er/images-of-the-day/sources"
	"github.com/spf13/cobra"
)

var (
	destinationDir   string
	sourceArgs       []string
	removeOtherFiles bool
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download images of the day from various sources",

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		run(ctx)
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

	downloadCmd.Flags().BoolVarP(
		&removeOtherFiles,
		"remove-other-files",
		"r",
		false,
		"Remove other files in the destination directory")
}

func run(ctx context.Context) {
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

	resultChannels := make([]chan sources.Result[sources.DownloadedImage], len(s))
	for i, source := range s {
		println("Downloading images from", source.GetName())
		resultChannel := sources.DownloadImages(source, ctx, destinationDir, date)
		resultChannels[i] = resultChannel
	}

	downloadedImagesChannel := make(chan sources.DownloadedImage, 10)

	wg := &sync.WaitGroup{}
	wg.Add(len(s))
	for i, resultChannel := range resultChannels {
		go func(i int, resultChannel chan sources.Result[sources.DownloadedImage]) {
			defer wg.Done()
			for result := range resultChannel {
				if result.Err != nil {
					println("Error in source:", s[i].GetName(), result.Err.Error())
					continue
				}
				println(result.Value.Message, result.Value.FilePath)
				downloadedImagesChannel <- result.Value
			}
		}(i, resultChannel)
	}

	downloadedImages := []sources.DownloadedImage{}
	go func() {
		for image := range downloadedImagesChannel {
			downloadedImages = append(downloadedImages, image)
		}
	}()

	wg.Wait()

	if removeOtherFiles {
		allFiles, err := os.ReadDir(destinationDir)
		if err != nil {
			println("Error reading directory:", err.Error())
			return
		}
		for _, file := range allFiles {
			if file.IsDir() {
				continue
			}
			fileName := file.Name()
			filePath := destinationDir + "/" + fileName

			found := false
			for _, downloadedImage := range downloadedImages {
				if filePath == downloadedImage.FilePath {
					found = true
					break
				}
			}
			if !found {
				println("Deleting file:", fileName)
				err := os.Remove(filePath)
				if err != nil {
					println("Error deleting file:", fileName, err.Error())
				} else {
					println("Deleted file:", fileName)
				}
			}
		}
	}
}
