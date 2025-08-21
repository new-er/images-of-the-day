package sources

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/gocolly/colly"
)

type Nasa struct {
}

func (u Nasa) GetName() string {
	return "Nasa"
}

func (u Nasa) GetImageLinks(ctx context.Context) chan ChannelResult[ImageLink] {
	c := newCollector()
	results := make(chan ChannelResult[ImageLink], 10)

	c.OnResponse(func(r *colly.Response) {
		nasaResponse := nasaResponse{}
		err := xml.Unmarshal(r.Body, &nasaResponse)
		if err != nil {
			select {
			case results <- ChannelResult[ImageLink]{Err: fmt.Errorf("failed to unmarshal NASA response: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		for _, item := range nasaResponse.Channel.Items {
			pubDate, err := time.Parse("Mon, 02 Jan 2006 15:04 MST", item.PubDate)
			if err != nil {
				results <- ChannelResult[ImageLink]{Err: fmt.Errorf("failed to parse pubDate: %w", err)}
				continue
			}
			if pubDate.Before(time.Now().Add(-48 * time.Hour)) {
				continue
			}
			results <- ChannelResult[ImageLink]{Value: ImageLink{URL: item.Enclosure.URL, Description: item.Title + " - " + item.Description}}
		}
	})

	go func() {
		defer close(results)
		err := c.Visit("https://www.nasa.gov/rss/dyn/lg_image_of_the_day.rss")

		if err != nil {
			select {
			case results <- ChannelResult[ImageLink]{Err: fmt.Errorf("failed to visit NASA RSS feed: %w", err)}:
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
	Enclosure   enclosure `xml:"enclosure"`
	PubDate     string    `xml:"pubDate"`
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
}

type enclosure struct {
	URL string `xml:"url,attr"`
}
