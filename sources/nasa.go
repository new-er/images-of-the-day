package sources

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gocolly/colly"
)

type Nasa struct {
}

func (u Nasa) GetImageLinks() ([]string, error) {
	c := colly.NewCollector()
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

func (u Nasa) SaveImages(destination string) error {
	url := "https://www.nasa.gov/rss/dyn/lg_image_of_the_day.rss"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	nasaResponse := nasaResponse{}
	err = xml.NewDecoder(resp.Body).Decode(&nasaResponse)
	if err != nil {
		return err
	}

	downloadErrors := []error{}
	for i, item := range nasaResponse.Channel.Items {
		file, err := os.Create(fmt.Sprintf("%s/nasa_%d.jpg", destination, i))
		if err != nil {
			downloadErrors = append(downloadErrors, err)
			continue
		}

		resp, err := http.Get(item.Enclosure.URL)
		if err != nil {
			downloadErrors = append(downloadErrors, err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			downloadErrors = append(downloadErrors, errors.New("Error"))
			continue
		}

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			downloadErrors = append(downloadErrors, err)
			continue
		}
	}

	return nil
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
