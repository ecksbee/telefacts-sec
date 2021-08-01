package names

import (
	"encoding/json"
	"io/ioutil"
)

func UnmarshalNames() (map[string]map[string]string, error) {
	names := make(map[string]map[string]string)
	b, err := ioutil.ReadFile("/names.json")
	if err != nil {
		return names, err
	}
	err = json.Unmarshal(b, &names)
	return names, err
}
