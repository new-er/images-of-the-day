package sources

import (
	"encoding/xml"
	"errors"
	"time"

	"github.com/gocolly/colly"
)

type Nasa struct {
}

func (u Nasa) GetName() string {
	return "Nasa"
}

func (u Nasa) GetImageLinks() ([]string, error) {
	c := newCollector()
	var imageLinks []string
	var errs []error

	c.OnResponse(func(r *colly.Response) {
		nasaResponse := nasaResponse{}
		err := xml.Unmarshal(r.Body, &nasaResponse)
		if err != nil {
			errs = append(errs, err)
			return
		}
		for _, item := range nasaResponse.Channel.Items {
			imageLinks = append(imageLinks, item.Enclosure.URL)
		}
	})

	err := c.Visit("https://www.nasa.gov/rss/dyn/lg_image_of_the_day.rss")
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return imageLinks, nil
}

type nasaResponse struct {
	Channel channel `xml:"channel"`
}

type channel struct {
	Items []item `xml:"item"`
}

type item struct {
	Enclosure enclosure `xml:"enclosure"`
}

type enclosure struct {
	URL string `xml:"url,attr"`
}
