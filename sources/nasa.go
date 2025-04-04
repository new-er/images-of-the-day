package sources

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/gocolly/colly"
)

type Nasa struct {
}

func (u Nasa) GetName() string {
	return "Nasa"
}

func (u Nasa) GetImageLinks(ctx context.Context) chan Result[string] {
	c := newCollector()
	results := make(chan Result[string], 10)

	c.OnResponse(func(r *colly.Response) {
		nasaResponse := nasaResponse{}
		err := xml.Unmarshal(r.Body, &nasaResponse)
		if err != nil {
			select {
			case results <- Result[string]{Err: fmt.Errorf("failed to unmarshal NASA response: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		for _, item := range nasaResponse.Channel.Items {
			select {
			case results <- Result[string]{Value: item.Enclosure.URL}:
			case <-ctx.Done():
				return
			}
		}
	})

	go func() {
		defer close(results)
		err := c.Visit("https://www.nasa.gov/rss/dyn/lg_image_of_the_day.rss")

		if err != nil {
			select {
			case results <- Result[string]{Err: fmt.Errorf("failed to visit NASA RSS feed: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
	}()
	return results
}

type nasaResponse struct {
	Channel channel `xml:"channel"`
}

type channel struct {
	Items []item `xml:"item"`
}

type item struct {
	Enclosure enclosure `xml:"enclosure"`
}

type enclosure struct {
	URL string `xml:"url,attr"`
}
