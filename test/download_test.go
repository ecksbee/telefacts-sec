package telefacts_sec_test

import (
	"os"
	"path/filepath"
	"testing"

	"ecksbee.com/telefacts-sec/pkg/serializables"
	"ecksbee.com/telefacts-sec/pkg/throttle"
)

func TestAllDownloads(t *testing.T) {
	throttle.StartSECThrottle()
	testDownloadGoFiler(t)
	testDownloadThunderDome(t)
	testDownloadWDesk(t)
	testDownloadWDesk2(t)
	testDownloadImages(t)
	testDownloadWDesk3(t)
}

func testDownloadGoFiler(t *testing.T) {
	dir, _ := os.Getwd()
	wd := filepath.Join(dir, "wd")
	gts := filepath.Join(dir, "gts")
	_, err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/843006/000165495420001999",
		wd, gts, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testDownloadThunderDome(t *testing.T) {
	dir, _ := os.Getwd()
	wd := filepath.Join(dir, "wd")
	gts := filepath.Join(dir, "gts")
	_, err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/69891/000143774920014395",
		wd, gts, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testDownloadWDesk(t *testing.T) {
	dir, _ := os.Getwd()
	wd := filepath.Join(dir, "wd")
	gts := filepath.Join(dir, "gts")
	_, err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/1445305/000144530520000124",
		wd, gts, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testDownloadWDesk2(t *testing.T) {
	dir, _ := os.Getwd()
	wd := filepath.Join(dir, "wd")
	gts := filepath.Join(dir, "gts")
	_, err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/0001058090/000105809020000020",
		wd, gts, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testDownloadWDesk3(t *testing.T) {
	dir, _ := os.Getwd()
	wd := filepath.Join(dir, "wd")
	gts := filepath.Join(dir, "gts")
	_, err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/1445305/000144530522000110",
		wd, gts, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testDownloadImages(t *testing.T) {
	dir, _ := os.Getwd()
	wd := filepath.Join(dir, "wd")
	gts := filepath.Join(dir, "gts")
	_, err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/320193/000032019322000006",
		wd, gts, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
