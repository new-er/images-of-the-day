package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

func DownloadImages(source Source, ctx context.Context, destinationDir, imagePrefix string) chan Result[string] {
	downloadedImages := []string{}
	downloadImagesResultChan := make(chan Result[string], 10)

	getImageLinksResultChan := source.GetImageLinks(ctx)

	go func() {
		defer close(downloadImagesResultChan)

		for result := range getImageLinksResultChan {
			if result.Err != nil {
				downloadImagesResultChan <- Result[string]{Err: fmt.Errorf("error getting image link: %w", result.Err)}
				continue
			}
			if slices.Contains(downloadedImages, result.Value) {
				continue
			}
			downloadedImages = append(downloadedImages, result.Value)

			linkParts := strings.Split(result.Value, "/")
			fileName := linkParts[len(linkParts)-1]
			fileName = strings.ReplaceAll(fileName, ":", "_")
			fileName = strings.ReplaceAll(fileName, "?", "_")
			fileName = strings.ReplaceAll(fileName, "&", "_")
			fileName = strings.ReplaceAll(fileName, "=", "_")
			fileName = strings.ReplaceAll(fileName, " ", "_")
			fileName = strings.ReplaceAll(fileName, ".", "_")
			fileName = strings.ReplaceAll(fileName, "jpg", "_jpg")
			fileName = strings.ReplaceAll(fileName, "jpeg", "_jpeg")
			fileName = strings.ReplaceAll(fileName, "png", "_png")
			fileName = strings.ReplaceAll(fileName, "gif", "_gif")
			fileName =	fmt.Sprintf("%s_%s_%s.jpg", imagePrefix, source.GetName(), fileName)
			filePath := fmt.Sprintf("%s/%s", destinationDir, fileName)
			if _, err := os.Stat(filePath); err == nil {
				downloadImagesResultChan <- Result[string]{Value: fmt.Sprintf("image already exists: %s", filePath)}
				continue
			}
			err := downloadImage(result.Value, filePath)

			if err != nil {
				downloadImagesResultChan <- Result[string]{Err: fmt.Errorf("error downloading image: %w", err)}
			} else {
				downloadImagesResultChan <- Result[string]{Value: fmt.Sprintf("downloaded image %s to %s", result.Value, filePath)}
			}
			time.Sleep(2 * time.Second)
		}
	}()
	return downloadImagesResultChan
}

func downloadImage(link string, destination string) error {
	response, err := http.Get(link)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	f, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, response.Body)
	if err != nil {
		return err
	}
	return nil
}
