package sources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gocolly/colly"
)

type Bing struct {
}

func (b Bing) GetName() string {
	return "Bing"
}

func (b Bing) GetImageLinks(ctx context.Context) chan Result[string] {
	c := newCollector()
	results := make(chan Result[string], 10)

	c.OnResponse(func(r *colly.Response) {
		bingResponse := bingResponse{}
		err := json.Unmarshal(r.Body, &bingResponse)
		if err != nil {
			select {
			case results <- Result[string]{Err: fmt.Errorf("failed to unmarshal Bing response: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
		for _, image := range bingResponse.Images {
			select {
			case results <- Result[string]{Value: "https://www.bing.com" + image.URL}:
			case <-ctx.Done():
				return
			}
		}
	})

	go func() {
		defer close(results)
		err := c.Visit("https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1")
		if err != nil {
			select {
			case results <- Result[string]{Err: fmt.Errorf("failed to visit Bing: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
	}()
	return results
}

type bingResponse struct {
	Images []bingImage `json:"images"`
}

type bingImage struct {
	URL string `json:"url"`
}
