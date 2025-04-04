package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

func (a Apod) GetImageLinks(ctx context.Context) chan Result[string] {
	c := newCollector()
	results := make(chan Result[string], 10)

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
				select {
				case results <- Result[string]{Value: link}:
				case <-ctx.Done():
					return
				}
			})
		})
	})

	go func() {
		defer close(results)

		err := c.Visit("https://apod.nasa.gov/apod/astropix.html")
		if err != nil {
			select {
			case results <- Result[string]{Err: fmt.Errorf("failed to visit Apod: %w", err)}:
			case <-ctx.Done():
			}
			return
		}
	}()
	return results
}

func (a Apod) SaveImages(destination string) error {
	url := "https://apod.nasa.gov/apod/astropix.html"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bodyString := string(body)

	expression := regexp.MustCompile("source src=\".*\"")
	tagText := expression.FindString(bodyString)
	srcText := tagText[12 : len(tagText)-1]

	url = "https://apod.nasa.gov/apod/" + srcText
	respImg, errImg := http.Get(url)
	if errImg != nil {
		return errImg
	}
	defer resp.Body.Close()

	file, err := os.Create(destination + "/apod.mp4")
	if err != nil {
		return err
	}

	_, err = io.Copy(file, respImg.Body)
	if err != nil {
		return err
	}

	return nil
}
