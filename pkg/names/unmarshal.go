package names

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

var (
	namePath = path.Join(".", "data", "/names.json")
)

func UnmarshalNames() (map[string]map[string]string, error) {
	names := make(map[string]map[string]string)
	if _, err := os.Stat(namePath); os.IsNotExist(err) {
		return names, nil
	}
	b, err := ioutil.ReadFile(namePath)
	if err != nil {
		return names, err
	}
	err = json.Unmarshal(b, &names)
	return names, err
}
