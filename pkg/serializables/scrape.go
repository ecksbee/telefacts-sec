package serializables

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"sync"

	"ecksbee.com/telefacts-sec/internal/actions"
	"golang.org/x/net/html/charset"
)

func Scrape(filingURL string, throttle func(string)) ([]byte, error) {
	isSEC, _ := regexp.MatchString(regexSEC, filingURL)
	if !isSEC {
		return nil, fmt.Errorf("not an acceptable SEC address, " + filingURL)
	}
	body, err := actions.Scrape(filingURL+"/FilingSummary.xml", throttle)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(body)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	filingSummary := FilingSummary{}
	err = decoder.Decode(&filingSummary)
	if len(filingSummary.InputFiles) <= 0 || len(filingSummary.InputFiles[0].File) <= 0 || err != nil {
		return nil, fmt.Errorf("empty filing at "+filingURL+". %s\n\n%v", string(body), err)
	}
	instance := filingSummary.GetInstance()
	srcDoc := filingSummary.GetIxbrl()
	schemaName := filingSummary.GetSchema()
	preItem := filingSummary.GetPresentationLinkbase()
	defItem := filingSummary.GetDefinitionLinkbase()
	calItem := filingSummary.GetCalculationLinkbase()
	labItem := filingSummary.GetLabelLinkbase()
	imageItems := filingSummary.GetImages()
	var wg sync.WaitGroup
	wg.Add(6)
	var (
		schema        []byte
		instanceBytes []byte
		srcBytes      []byte
		presentation  []byte
		definition    []byte
		calculation   []byte
		label         []byte
	)
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + schemaName
		schema, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + instance
		instanceBytes, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + srcDoc
		srcBytes, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + preItem
		presentation, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + defItem
		definition, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + calItem
		calculation, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + labItem
		label, err = actions.Scrape(targetUrl, throttle)
	}()
	wg.Wait()
	if err != nil {
		return nil, err
	}
	images := []struct {
		Name string
		Body []byte
	}{}
	for _, imageItem := range imageItems {
		targetUrl := filingURL + "/" + imageItem
		image, err := actions.Scrape(targetUrl, throttle)
		if err == nil {
			images = append(images, struct {
				Name string
				Body []byte
			}{
				Name: imageItem,
				Body: image,
			})
		}
	}
	return zipData(append([]struct {
		Name string
		Body []byte
	}{
		{schemaName, schema},
		{instance, instanceBytes},
		{srcDoc, srcBytes},
		{preItem, presentation},
		{defItem, definition},
		{calItem, calculation},
		{labItem, label},
	}, images...))
}

func zipData(files []struct {
	Name string
	Body []byte
}) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	for _, file := range files {
		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return nil, err
		}
		_, err = zipFile.Write(file.Body)
		if err != nil {
			return nil, err
		}
	}
	err := zipWriter.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
