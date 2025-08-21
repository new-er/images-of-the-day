package sources

import (
	"context"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Epod struct {
}

func (e Epod) GetName() string {
	return "Epod"
}

func (e Epod) GetImageLinks(ctx context.Context) chan ChannelResult[ImageLink] {
	c := newCollector()
	results := make(chan ChannelResult[ImageLink], 10)

	c.OnHTML("div", func(e *colly.HTMLElement) {
		if e.Attr("class") != "entry-body" {
			return
		}
		url := ""
		description := ""

		e.DOM.Find("a").Each(func(i int, s *goquery.Selection) {
			value, exists := s.Attr("class")
			if !exists {
				return
			}
			if value != "asset-img-link" {
				return
			}
			href, hrefExists := s.Attr("href")
			if !hrefExists {
				return
			}
			url = href

			if description == "" {
				return
			}

			select {
			case results <- ChannelResult[ImageLink]{Value: ImageLink{URL: url, Description: description}}:
			case <-ctx.Done():
				return
			}
		})

		pIndex := -1
		e.DOM.ChildrenFiltered("p").Each(func(i int, s *goquery.Selection) {
			pIndex++
			if pIndex < 2 {
				return
			}

			if pIndex == 2 {
				description += s.Text()
				return
			}
			if pIndex == 3 {
				description += "\n" + s.Text()
				return
			}

			if url == "" {
				return
			}

			select {
			case results <- ChannelResult[ImageLink]{Value: ImageLink{URL: url, Description: description}}:
			case <-ctx.Done():
				return
			}
		})
	})

	go func() {
		defer close(results)
		err := c.Visit("https://epod.usra.edu/")
		if err != nil {
			select {
			case results <- ChannelResult[ImageLink]{Err: fmt.Errorf("failed to visit Epod: %w", err)}:
			case <-ctx.Done():
			}
		}
	}()

	return results
}
