package sources

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Epod struct {
}

func (e Epod) SaveImages(destination string) error {
	url := "https://epod.usra.edu/"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	expression := regexp.MustCompile(`<a\s+[^>]*?href\s*=\s*"([^"]+)"`)
	matches := expression.FindAllStringSubmatch(string(body), -1)

	testFile, err := os.Create("test.txt")
	if err != nil {
		return err
	}
	defer testFile.Close()
	testFile.WriteString(string(body))

	imageUrls := make([]string, 0)
	for _, match := range matches {
		if strings.Contains(match[0], "asset-img-link") {
			for _, imageUrl := range imageUrls {
				if imageUrl == match[1] {
					continue
				}
			}
			imageUrls = append(imageUrls, match[1])
		}
	}

	for i, imageUrl := range imageUrls {
		resp, err := http.Get(imageUrl)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return err
		}

		file, err := os.Create(fmt.Sprintf("%s/epod_%d.jpg", destination, i))
		if err != nil {
			return err
		}

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return err
		}
	}

	return nil
}
