package main

import (
	"images-of-the-day/sources"
)

func main() {
	sources := []sources.Source{
		sources.Bing{},
		sources.Nasa{},
		sources.Apod{},
		sources.EarthObservatory{},
		//sources.Epod{},
	}

	destination := "/home/neuer/Downloads/imagesoftheday"

	for _, source := range sources {
		err := source.SaveImages(destination)
		if err != nil {
			print(err.Error())
		}
	}
}
