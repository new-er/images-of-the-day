package sources

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Apod struct {
}

func (a Apod) GetName() string {
	return "Apod"
}

func (a Apod) GetImageLinks(ctx context.Context) chan ChannelResult[ImageLink] {
	c := newCollector()
	results := make(chan ChannelResult[ImageLink], 10)

	url := ""
	description := ""

	c.OnHTML("center", func(e *colly.HTMLElement) {
		hasHeader := false
		e.DOM.ChildrenFiltered("h1").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "Astronomy Picture of the Day") {
				hasHeader = true
			}
		})
		if !hasHeader {
			return
		}

		e.DOM.ChildrenFiltered("p").Each(func(i int, s *goquery.Selection) {
			s.ChildrenFiltered("a").Each(func(i int, s *goquery.Selection) {
				link, existsLink := s.Attr("href")
				if !existsLink {
					return
				}
				if !strings.Contains(link, "jpg") && !strings.Contains(link, "png") {
					return
				}
				httpRegExp := regexp.MustCompile(`^http`)

				if link[0] == '/' || !httpRegExp.MatchString(`^http`) {
					link = "https://apod.nasa.gov/apod/" + link
				}
				url = link
			})
		})

		if url == "" {
			return
		}
		if description == "" {
			return
		}

		select {
		case results <- ChannelResult[ImageLink]{
			Value: ImageLink{
				URL:         url,
				Description: description,
			},
		}:
		case <-ctx.Done():
			return
		}
	})

	c.OnHTML("p", func(e *colly.HTMLElement) {
		hasExplanation := false

		e.DOM.ChildrenFiltered("b").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "Explanation:") {
				hasExplanation = true
			}
		})

		if !hasExplanation {
			return
		}

		description = e.DOM.Text()

		if url == "" {
			return
		}
		if description == "" {
			return
		}

		select {
		case results <- ChannelResult[ImageLink]{
			Value: ImageLink{
				URL:         url,
				Description: description,
			},
		}:
		case <-ctx.Done():
			return
		}
	})

	go func() {
		defer close(results)

		err := c.Visit("https://apod.nasa.gov/apod/astropix.html")
		if err != nil {
			select {
			case results <- ChannelResult[ImageLink]{Err: fmt.Errorf("failed to visit Apod: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
	}()
	return results
}
