package sources

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"slices"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type EarthObservatory struct {
}

func (e EarthObservatory) GetName() string {
	return "EarthObservatory"
}

func (e EarthObservatory) GetImageLinks(ctx context.Context) chan ChannelResult[ImageLink] {
	c := newCollector()
	imageLinksSlice := []string{}
	results := make(chan ChannelResult[ImageLink], 10)

	c.OnResponse(func(r *colly.Response) {
		earthObservatoryResponse := earthObservatoryResponse{}
		err := xml.Unmarshal(r.Body, &earthObservatoryResponse)
		if err != nil {
			select {
			case results <- ChannelResult[ImageLink]{Err: fmt.Errorf("failed to unmarshal Earth Observatory response: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		for _, item := range earthObservatoryResponse.Channel.Items {
			pubDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
			if err != nil {
				results <- ChannelResult[ImageLink]{Err: fmt.Errorf("failed to parse pubDate: %w", err)}
				continue
			}
			if pubDate.Before(time.Now().Add(-48 * time.Hour)) {
				continue
			}
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
							case results <- ChannelResult[ImageLink]{Err: errors.New("href attribute not found")}:
							case <-ctx.Done():
							}
							return
						}

						if slices.Contains(imageLinksSlice, val) {
							return
						}
						select {
						case results <- ChannelResult[ImageLink]{Value: ImageLink{URL: val, Description: item.Title + " - " + item.Description}}:
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
			case results <- ChannelResult[ImageLink]{Err: fmt.Errorf("failed to visit Earth Observatory RSS feed: %w", err)}:
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
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
}
