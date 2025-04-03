package sources

import (
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

func (a Apod) GetImageLinks() ([]string, error) {
	c := newCollector()
	l := []string{}
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
				l = append(l, link)
			})
		})
	})

	err := c.Visit("https://apod.nasa.gov/apod/astropix.html")
	if err != nil {
		return nil, err
	}
	return l, nil
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
