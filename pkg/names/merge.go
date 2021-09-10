package names

import (
	"bytes"
	"encoding/json"

	"ecksbee.com/telefacts-sec/internal/actions"
)

func MergeNames(throttle func(string)) error {
	unmarshalled, err := UnmarshalNames()
	if err != nil {
		return err
	}
	scraped, err := ScrapeNames(throttle)
	if err != nil {
		return err
	}
	merged := unmarshalled
	for scrapedScheme, scrapedSchemeMap := range scraped {
		for scrapedChardata, scrapedName := range scrapedSchemeMap {
			if _, found := merged[scrapedScheme]; !found {
				merged[scrapedScheme] = make(map[string]string)
			}
			merged[scrapedScheme][scrapedChardata] = scrapedName
		}
	}
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(merged)
	return actions.WriteFile(namePath, buffer.Bytes())
}
