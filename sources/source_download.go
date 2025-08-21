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

func DownloadImages(source Source, ctx context.Context, destinationDir, imagePrefix string) chan Result[DownloadedImage] {
	downloadedImages := []string{}
	downloadImagesResultChan := make(chan Result[DownloadedImage], 10)

	getImageLinksResultChan := source.GetImageLinks(ctx)

	go func() {
		defer close(downloadImagesResultChan)

		for result := range getImageLinksResultChan {
			if result.Err != nil {
				downloadImagesResultChan <- Result[DownloadedImage]{Err: fmt.Errorf("error getting image link: %w", result.Err)}
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
			fileName = fmt.Sprintf("%s_%s_%s.jpg", source.GetName(), imagePrefix, fileName)
			filePath := fmt.Sprintf("%s/%s", destinationDir, fileName)

			if _, err := os.Stat(filePath); err == nil {
				downloadedImage := DownloadedImage{
					ImageLink: result.Value,
					FilePath:  filePath,
					Message:   fmt.Sprintf("image already exists: %s", filePath),
				}
				downloadImagesResultChan <- Result[DownloadedImage]{Value: downloadedImage}
				continue
			}

			err := downloadImage(result.Value, filePath)

			if err != nil {
				downloadImagesResultChan <- Result[DownloadedImage]{Err: fmt.Errorf("error downloading image: %w", err)}
			} else {
				downloadedImage := DownloadedImage{
					ImageLink: result.Value,
					FilePath:  filePath,
					Message:   fmt.Sprintf("image downloaded: %s", filePath),
				}
				downloadImagesResultChan <- Result[DownloadedImage]{Value: downloadedImage}
				time.Sleep(2 * time.Second)
			}
		}
	}()
	return downloadImagesResultChan
}

type DownloadedImage struct {
	ImageLink string
	FilePath  string
	Message   string
}

func downloadImage(link string, destination string) error {
	client := http.Client{
		Transport: &http.Transport{},
	}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	f, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}
	return nil
}
