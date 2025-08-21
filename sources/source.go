package sources

import (
	"context"
	"time"

	"github.com/gocolly/colly"
)

type ImageLink struct {
	URL	string
	Description string
}

type Source interface {
	GetName() string
	GetImageLinks(ctx context.Context) chan ChannelResult[ImageLink]
}

func newCollector() *colly.Collector {
	c := colly.NewCollector()
	c.Limit(&colly.LimitRule{
		DomainGlob: "*",
		Delay:      5 * time.Second,
	})
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	return c
}
