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

func (a Apod) GetImageLinks(ctx context.Context) chan Result[ImageDescription] {
	c := newCollector()
	results := make(chan Result[ImageDescription], 10)

	url := ""
	title := ""

	c.OnHTML("center", func(e *colly.HTMLElement) {
		hasHeader := false
		hasTitle := false
		e.DOM.ChildrenFiltered("h1").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "Astronomy Picture of the Day") {
				hasHeader = true
			}
		})

		e.DOM.ChildrenFiltered("b").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "Image Credit") {
				hasTitle = true
			}
		})

		if hasHeader {
			getImageUrl(e, func(s string) {
				url = s
				emitOnAllValuesFound(ctx, url, title, "https://apod.nasa.gov/apod/astropix.html", results)
			})
		}

		if hasTitle {
			title = getTitle(e)
			emitOnAllValuesFound(ctx, url, title, "https://apod.nasa.gov/apod/astropix.html", results)
		}
	})

	go func() {
		defer close(results)

		err := c.Visit("https://apod.nasa.gov/apod/astropix.html")
		if err != nil {
			select {
			case results <- Result[ImageDescription]{Err: fmt.Errorf("failed to visit Apod: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
	}()
	return results
}

func emitOnAllValuesFound(ctx context.Context, imageUrl, title, pageUrl string, results chan Result[ImageDescription]) {
	if imageUrl == "" {
		return
	}
	if title == "" {
		return
	}
	select {
	case results <- Result[ImageDescription]{Value: ImageDescription{
		ImageUrl: imageUrl, Title: title, PageUrl: "https://apod.nasa.gov/apod/astropix.html"}}:
	case <-ctx.Done():
		return
	}
}

func getImageUrl(e *colly.HTMLElement, f func(string)) {
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
			f(link)
		})
	})
}

func getTitle(e *colly.HTMLElement) string {
	element := e.DOM.ChildrenFiltered("b").Get(0)
	return strings.Trim(element.FirstChild.Data, " ")
}
