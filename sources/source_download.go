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

func DownloadImages(source Source, ctx context.Context, destinationDir, imagePrefix string) chan ChannelResult[DownloadedImage] {
	downloadedImages := []string{}
	downloadImagesResultChan := make(chan ChannelResult[DownloadedImage], 10)

	getImageLinksResultChan := source.GetImageLinks(ctx)

	go func() {
		defer close(downloadImagesResultChan)

		for result := range getImageLinksResultChan {
			if result.Err != nil {
				downloadImagesResultChan <- ChannelResult[DownloadedImage]{Err: fmt.Errorf("error getting image link: %w", result.Err)}
				continue
			}
			if slices.Contains(downloadedImages, result.Value.URL) {
				continue
			}
			downloadedImages = append(downloadedImages, result.Value.URL)

			linkParts := strings.Split(result.Value.URL, "/")
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
			fileName = fmt.Sprintf("%s_%s_%s.jpg", imagePrefix, source.GetName(), fileName)
			filePath := fmt.Sprintf("%s/%s", destinationDir, fileName)

			if _, err := os.Stat(filePath); err == nil {
				downloadedImage := DownloadedImage{
					ImageLink: result.Value.URL,
					FilePath:  filePath,
					Message:   fmt.Sprintf("image already exists: %s", filePath),
				}
				downloadImagesResultChan <- ChannelResult[DownloadedImage]{Value: downloadedImage}
				continue
			}

			err := downloadImage(result.Value.URL, filePath)

			if err != nil {
				downloadImagesResultChan <- ChannelResult[DownloadedImage]{Err: fmt.Errorf("error downloading image: %w", err)}
			}

			err = writeDescriptionFile(filePath+".txt", result.Value.Description)
			if err != nil {
				downloadImagesResultChan <- ChannelResult[DownloadedImage]{Err: fmt.Errorf("error writing description file: %w", err)}
			}

			downloadedImage := DownloadedImage{
				ImageLink: result.Value.URL,
				FilePath:  filePath,
				Message:   fmt.Sprintf("image downloaded: %s", filePath),
			}
			downloadImagesResultChan <- ChannelResult[DownloadedImage]{Value: downloadedImage}
			time.Sleep(2 * time.Second)
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

func writeDescriptionFile(destination string, description string) error {
	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create description file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(description)
	if err != nil {
		return fmt.Errorf("failed to write description: %w", err)
	}

	return nil
}
