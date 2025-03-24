package sources

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type EarthObservatory struct {
}

func (e EarthObservatory) SaveImages(destination string) error {
	url := "https://earthobservatory.nasa.gov/feeds/earth-observatory.rss"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	earthObservatoryResponse := earthObservatoryResponse{}
	err = xml.NewDecoder(resp.Body).Decode(&earthObservatoryResponse)
	if err != nil {
		return err
	}

	for i, item := range earthObservatoryResponse.Channel.Items {
		if err != nil {
			return err
		}

		resp, err := http.Get(item.Link)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		expression := regexp.MustCompile(`<a\s+[^>]*?href\s*=\s*"([^"]+)"`)
		matches := expression.FindAllStringSubmatch(string(body), -1)

		imageUrls := make([]string, 0)
		for _, match := range matches {
			if strings.Contains(match[1], "jpg") {
				for _, imageUrl := range imageUrls {
					if imageUrl == match[1] {
						continue
					}
				}
				imageUrls = append(imageUrls, match[1])
			}
		}

		for j, imageUrl := range imageUrls {
			file, err := os.Create(fmt.Sprintf("%s/earthobservatory_%d_%d.jpg", destination, i, j))
			if err != nil {
				return err
			}

			respImg, err := http.Get(imageUrl)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			_, err = io.Copy(file, respImg.Body)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type earthObservatoryResponse struct {
	Channel EarthObservatoryChannel `xml:"channel"`
}

type EarthObservatoryChannel struct {
	Items []EarthObservatoryItem `xml:"item"`
}

type EarthObservatoryItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}
