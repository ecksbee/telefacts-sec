package names

import (
	"bytes"
	"encoding/json"
	"os"
)

func WriteFile(dest string, data []byte) error {
	file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
func MergeNames() error {
	unmarshalled, err := UnmarshalNames()
	if err != nil {
		return err
	}
	scraped, err := ScrapeNames()
	if err != nil {
		return err
	}
	merged := unmarshalled
	for scrapedScheme, scrapedSchemeMap := range scraped {
		for scrapedChardata, scrapedName := range scrapedSchemeMap {
			merged[scrapedScheme][scrapedChardata] = scrapedName
		}
	}
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(merged)
	WriteFile("names.json", buffer.Bytes())
	return nil
}
