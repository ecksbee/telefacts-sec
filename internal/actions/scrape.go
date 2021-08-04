package actions

import (
	"bytes"
	"io"
	"net/http"
)

const USERAGENT = "ECKSBEE LLC admin@ecksbee.com"

func Scrape(url string, throttle func(string)) ([]byte, error) {
	throttle(url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", USERAGENT)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, resp.Body)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
