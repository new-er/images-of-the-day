package sources

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

type Bing struct { 

}

func (b Bing) SaveImages(destination string) error {
	url := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("Error")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bingResponse := bingResponse{}
	json.Unmarshal(body, &bingResponse)

	downloadErrors := []error{}
	for _, image := range bingResponse.Images {
		url := "https://www.bing.com" + image.URL
		file, err := os.Create(destination + "/bing.jpg")
		if err != nil {
			downloadErrors = append(downloadErrors, err)
			continue
		}

		resp, err := http.Get(url)
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
	if len(downloadErrors) > 0 {
		return errors.Join(downloadErrors...)
	}

	return nil

}

type bingResponse struct {
	Images []bingImage `json:"images"`
}

type bingImage struct {
	URL string `json:"url"`
}
