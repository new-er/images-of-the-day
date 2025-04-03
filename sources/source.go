package sources

import (
	"time"

	"github.com/gocolly/colly"
)

type Source interface {
	GetName() string
	GetImageLinks() ([]string, error)
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
