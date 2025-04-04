package sources

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"slices"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type EarthObservatory struct {
}

func (e EarthObservatory) GetName() string {
	return "EarthObservatory"
}

func (e EarthObservatory) GetImageLinks(ctx context.Context) chan Result[string] {
	c := newCollector()
	imageLinksSlice := []string{}
	results := make(chan Result[string], 10)

	c.OnResponse(func(r *colly.Response) {
		earthObservatoryResponse := earthObservatoryResponse{}
		err := xml.Unmarshal(r.Body, &earthObservatoryResponse)
		if err != nil {
			select {
			case results <- Result[string]{Err: fmt.Errorf("failed to unmarshal Earth Observatory response: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		for _, item := range earthObservatoryResponse.Channel.Items {
			c := newCollector()
			c.OnHTML("div", func(e *colly.HTMLElement) {
				if e.Attr("class") == "panel-image" {
					e.DOM.ChildrenFiltered("a").Each(func(i int, s *goquery.Selection) {
						targetV, existsTarget := s.Attr("target")
						if !existsTarget {
							return
						}
						if targetV != "_blank" {
							return
						}
						val, exists := s.Attr("href")
						if !exists {
							select {
							case results <- Result[string]{Err: errors.New("href attribute not found")}:
							case <-ctx.Done():
							}
							return
						}

						if slices.Contains(imageLinksSlice, val) {
							return
						}
						select {
						case results <- Result[string]{Value: val}:
						case <-ctx.Done():
							return
						}
						imageLinksSlice = append(imageLinksSlice, val)
					})
				}
			})
			c.Visit(item.Link)
		}
	})

	go func() {
		defer close(results)
		err := c.Visit("https://earthobservatory.nasa.gov/feeds/earth-observatory.rss")
		if err != nil {
			select {
			case results <- Result[string]{Err: fmt.Errorf("failed to visit Earth Observatory RSS feed: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
	}()
	return results
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
