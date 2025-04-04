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

func (e Epod) GetImageLinks(ctx context.Context) chan Result[string] {
	c := newCollector()
	results := make(chan Result[string], 10)

	c.OnHTML("a", func(e *colly.HTMLElement) {
		if e.Attr("class") == "asset-img-link" {
			href := e.Attr("href")
			select {
			case results <- Result[string]{Value: href}:
			case <-ctx.Done():
				return
			}
		}
	})

	go func() {
		defer close(results)
		err := c.Visit("https://epod.usra.edu/")
		if err != nil {
			select {
			case results <- Result[string]{Err: fmt.Errorf("failed to visit Epod: %w", err)}:
			case <-ctx.Done():
			}
		}
	}()

	return results
}
