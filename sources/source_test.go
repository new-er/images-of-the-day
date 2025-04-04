package sources

import (
	"context"
	"testing"
)

var sources = []Source{
	&Apod{},
	&Bing{},
	&EarthObservatory{},
	&Epod{},
	&Nasa{},
}

func TestSourceCanGetImageLinks(t *testing.T) {
	for _, source := range sources {
		t.Run(source.GetName(), func(t *testing.T) {
			items := []string{}
			itemsChan, errorChan := source.GetImageLinks(context.Background())
			select {
			case item := <-itemsChan:
				items = append(items, item)
			case err := <-errorChan:
				t.Errorf("Error: %v", err)
			}

			println("Items from", source.GetName())
			for _, item := range items {
				println(item)
			}
		})
	}
}
