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
}

func testDownloadGoFiler(t *testing.T) {
	wd, _ := os.Getwd()
	dir := filepath.Join(wd, "data")
	err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/843006/000165495420001999",
		dir, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testDownloadThunderDome(t *testing.T) {
	wd, _ := os.Getwd()
	dir := filepath.Join(wd, "data")
	err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/69891/000143774920014395",
		dir, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func testDownloadWDesk(t *testing.T) {
	wd, _ := os.Getwd()
	dir := filepath.Join(wd, "data")
	err := serializables.Download(
		"https://www.sec.gov/Archives/edgar/data/1445305/000144530520000124",
		dir, throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
