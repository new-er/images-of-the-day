package sources

import (
	"io"
	"net/http"
	"os"
	"regexp"
)

type Apod struct {
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
