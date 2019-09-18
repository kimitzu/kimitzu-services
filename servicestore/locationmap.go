package servicestore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func LoadLocationMap() map[string]map[string][]float64 {
	fStream, err := ioutil.ReadFile("./locationmap.json")
	if err != nil {
		fmt.Printf("Failed Reading file %v\n", err)
	}
	obj := make(map[string]map[string][]float64)
	json.Unmarshal(fStream, &obj)
	return obj
}
