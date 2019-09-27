package servicestore

import (
	"encoding/json"
	"fmt"

	"github.com/gobuffalo/packr/v2"
)

func LoadLocationMap() map[string]map[string][]float64 {
	box := packr.New("external", "../external")
	fStream, err := box.Find("locationmap.json")
	if err != nil {
		fmt.Printf("Failed Reading file %v\n", err)
	}
	obj := make(map[string]map[string][]float64)
	json.Unmarshal(fStream, &obj)
	return obj
}
