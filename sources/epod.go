package sources

import (
	"context"
	"fmt"

	"github.com/gocolly/colly"
)

type Epod struct {
}

func (e Epod) GetName() string {
	return "Epod"
}

func (e Epod) GetImageLinks(ctx context.Context) chan Result[ImageDescription] {
	c := newCollector()
	results := make(chan Result[ImageDescription], 10)

	imageUrl := ""
	title := ""

	c.OnHTML("a", func(e *colly.HTMLElement) {
		if e.Attr("class") == "asset-img-link" {
			imageUrl = e.Attr("href")
			emitOnAllValuesFound(ctx, imageUrl, title, "https://epod.usra.edu/", results)
		}
	})

	c.OnHTML("h3", func(h *colly.HTMLElement) {
		if h.Attr("class") != "entry-header" {
			return
		}

		title = h.DOM.Children().Get(0).FirstChild.Data
		emitOnAllValuesFound(ctx, imageUrl, title, "https://epod.usra.edu/", results)
	})

	go func() {
		defer close(results)
		err := c.Visit("https://epod.usra.edu/")
		if err != nil {
			select {
			case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to visit Epod: %w", err)}:
			case <-ctx.Done():
			}
		}
	}()

	return results
}
