package main

import (
	"images-of-the-day/sources"
)

func main() {
	sources := []sources.Source{
		sources.Bing{},
		//sources.Nasa{},
		sources.Apod{},
		sources.EarthObservatory{},
		sources.Epod{},
	}

	destination := "/home/neuer/Downloads/imagesoftheday"

	for _, source := range sources {
		links, err := source.GetImageLinks()
		if err != nil {
			println(err.Error())
		}
		for _, link := range links {
			println("Downloading " + link + " to " + destination)
		}
	}
}
