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

type EarthObservatory struct {
}

func (e EarthObservatory) GetName() string {
	return "EarthObservatory"
}

func (e EarthObservatory) GetImageLinks(ctx context.Context) chan Result[ImageDescription] {
	c := newCollector()
	imageLinksSlice := []string{}
	results := make(chan Result[ImageDescription], 10)

	c.OnResponse(func(r *colly.Response) {
		earthObservatoryResponse := earthObservatoryResponse{}
		err := xml.Unmarshal(r.Body, &earthObservatoryResponse)
		if err != nil {
			select {
			case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to unmarshal Earth Observatory response: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		for _, item := range earthObservatoryResponse.Channel.Items {
			pubDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
			if err != nil {
				select {
				case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to parse pubDate: %w", err)}:
				case <-ctx.Done():
				}
				return
			}
			if pubDate.Before(time.Now().Add(-7 * 24 * time.Hour)) {
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
				case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to visit Earth Observatory item: %w", err)}:
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
			case results <- Result[ImageDescription]{Value: ImageDescription{ImageUrl: imageUrl, Title: item.Title, PageUrl: item.Guid}}:
			case <-ctx.Done():
				return
			}
			imageLinksSlice = append(imageLinksSlice, imageUrl)
		}
	})

	go func() {
		defer close(results)
		err := c.Visit("https://science.nasa.gov/feed/earth-observatory/image-of-the-day")
		if err != nil {
			select {
			case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to visit Earth Observatory RSS feed: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
	}()
	return results
}

type earthObservatoryResponse struct {
	Channel EarthObservatoryChannel `xml:"channel"`
}

type EarthObservatoryChannel struct {
	Items []EarthObservatoryItem `xml:"item"`
}

type EarthObservatoryItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	Guid    string `xml:"guid"`
}
