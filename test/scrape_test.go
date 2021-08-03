package telefacts_sec_test

import (
	"os"
	"path"
	"testing"

	"ecksbee.com/telefacts-sec/pkg/serializables"
)

func TestAllScrapes(t *testing.T) {
	startSECThrottle()
	testScrape(t)
	testScrape_Large(t)
	testScrape_Gold(t)
}

func testScrape(t *testing.T) {
	workingDir := path.Join(".", "data")
	pathStr := path.Join(workingDir, "test_small")
	err := os.Mkdir(pathStr, 0755)
	if err != nil {
		t.Fatalf("Error: " + err.Error())
	}
	defer os.RemoveAll(pathStr)
	err = serializables.Scrape(
		"https://www.sec.gov/Archives/edgar/data/843006/000165495420001999",
		pathStr, throttle)
	if err != nil {
		t.Fatalf("Error: " + err.Error())
	}
}

func testScrape_Large(t *testing.T) {
	workingDir := path.Join(".", "data")
	pathStr := path.Join(workingDir, "test_large")
	err := os.Mkdir(pathStr, 0755)
	if err != nil {
		t.Fatalf("Error: " + err.Error())
	}
	defer os.RemoveAll(pathStr)
	err = serializables.Scrape(
		"https://www.sec.gov/Archives/edgar/data/69891/000143774920014395",
		pathStr, throttle)
	if err != nil {
		t.Fatalf("Error: " + err.Error())
	}
}

func testScrape_Gold(t *testing.T) {
	workingDir := path.Join(".", "data")
	pathStr := path.Join(workingDir, "test_gold")
	err := os.Mkdir(pathStr, 0755)
	if err != nil {
		t.Fatalf("Error: " + err.Error())
	}
	defer os.RemoveAll(pathStr)
	err = serializables.Scrape(
		"https://www.sec.gov/Archives/edgar/data/1445305/000144530520000124",
		pathStr, throttle)
	if err != nil {
		t.Fatalf("Error: " + err.Error())
	}
}
