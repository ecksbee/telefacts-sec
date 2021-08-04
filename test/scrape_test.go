package telefacts_sec_test

import (
	"path"
	"testing"

	"ecksbee.com/telefacts-sec/pkg/serializables"
)

var (
	dir = path.Join(".", "data")
)

func TestAllScrapes(t *testing.T) {
	startSECThrottle()
	testScrapeGoFiler(t)
	testScrapeThunderDome(t)
	testScrapeWDesk(t)
}

func testScrapeGoFiler(t *testing.T) {
	err := serializables.Scrape(
		"https://www.sec.gov/Archives/edgar/data/843006/000165495420001999",
		dir, throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testScrapeThunderDome(t *testing.T) {
	err := serializables.Scrape(
		"https://www.sec.gov/Archives/edgar/data/69891/000143774920014395",
		dir, throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testScrapeWDesk(t *testing.T) {
	err := serializables.Scrape(
		"https://www.sec.gov/Archives/edgar/data/1445305/000144530520000124",
		dir, throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
