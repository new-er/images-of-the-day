package main

import (
	"fmt"
	"images-of-the-day/sources"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	sources := []sources.Source{
		sources.Bing{},
		sources.Nasa{},
		sources.Apod{},
		sources.EarthObservatory{},
		sources.Epod{},
	}

	destination := "/home/neuer/Downloads/imagesoftheday"
	date := time.Now().Format("2006-01-02")

	for _, source := range sources {
		links, err := source.GetImageLinks()
		if err != nil {
			println(err.Error())
		}
		for i, link := range links {
			response, err := http.Get(link)
			if err != nil {
				println(err.Error())
				continue
			}
			defer response.Body.Close()

			fileName := fmt.Sprintf("%s/%s_%s_%d.jpg", destination, date, source.GetName(), i)
			f, err := os.Create(fileName)
			if err != nil {
				println(err.Error())
				continue
			}
			defer f.Close()

			_, err = io.Copy(f, response.Body)
			if err != nil {
				println(err.Error())
				continue
			}
			println("Downloaded", fileName)
		}
	}
}
