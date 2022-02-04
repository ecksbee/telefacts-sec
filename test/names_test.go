package telefacts_sec_test

import (
	"path"
	"testing"

	"ecksbee.com/telefacts-sec/pkg/names"
	"ecksbee.com/telefacts-sec/pkg/throttle"
)

func TestMergeNames(t *testing.T) {
	throttle.StartSECThrottle()
	names.NamePath = path.Join(".", "data", "/names.json")
	err := names.MergeNames(throttle.Throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
