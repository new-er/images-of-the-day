package sources

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/gocolly/colly"
)

type Apod struct {
}

func (a Apod) GetName() string {
	return "Apod"
}

func (a Apod) GetImageLinks(ctx context.Context) chan Result[ImageDescription] {
	c := newCollector()
	imageLinksSlice := []string{}
	results := make(chan Result[ImageDescription], 10)

	c.OnResponse(func(r *colly.Response) {
		apodResponse := apodResponse{}
		err := xml.Unmarshal(r.Body, &apodResponse)
		if err != nil {
			select {
			case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to unmarshal Apod response: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		for _, item := range apodResponse.Channel.Items {
			pubDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
			if err != nil {
				select {
				case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to parse pubDate: %w", err)}:
				case <-ctx.Done():
				}
				return
			}
			if pubDate.Before(time.Now().Add(-48 * time.Hour)) {
				continue
			}

			imageUrl := ""
			ic := newCollector()
			ic.OnHTML("meta", func(e *colly.HTMLElement) {
				if e.Attr("property") == "og:image" && imageUrl == "" {
					imageUrl = e.Attr("content")
				}
			})
			err = ic.Visit(item.Link)
			if err != nil {
				select {
				case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to visit Apod item: %w", err)}:
				case <-ctx.Done():
				}
				return
			}
			if imageUrl == "" {
				select {
				case results <- Result[ImageDescription]{Err: errors.New("og:image not found")}:
				case <-ctx.Done():
				}
				return
			}

			if slices.Contains(imageLinksSlice, imageUrl) {
				continue
			}
			select {
			case results <- Result[ImageDescription]{Value: ImageDescription{ImageUrl: imageUrl, Title: item.Title, PageUrl: item.Link}}:
			case <-ctx.Done():
				return
			}
			imageLinksSlice = append(imageLinksSlice, imageUrl)
		}
	})

	go func() {
		defer close(results)
		err := c.Visit("https://science.nasa.gov/feed/apod-basic/")
		if err != nil {
			select {
			case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to visit Apod feed: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
	}()
	return results
}

type apodResponse struct {
	Channel apodChannel `xml:"channel"`
}

type apodChannel struct {
	Items []apodItem `xml:"item"`
}

type apodItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

// emitOnAllValuesFound is kept for the Epod source, which still references it.
func emitOnAllValuesFound(ctx context.Context, imageUrl, title, pageUrl string, results chan Result[ImageDescription]) {
	if imageUrl == "" {
		return
	}
	if title == "" {
		return
	}
	select {
	case results <- Result[ImageDescription]{Value: ImageDescription{
		ImageUrl: imageUrl, Title: title, PageUrl: pageUrl}}:
	case <-ctx.Done():
		return
	}
}
