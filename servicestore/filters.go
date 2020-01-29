package servicestore

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/PaesslerAG/gval"

	"github.com/kimitzu/kimitzu-services/location"
	"github.com/kimitzu/kimitzu-services/models"
)

func any(arr []bool) bool {
	for _, x := range arr {
		if x {
			return x
		}
	}
	return false
}

func has(token, str string) bool {
	// def has(token, str):
	//...     return any(x.lower().startswith(token.lower()) for x in str.split(" "))
	tokens := strings.Split(str, " ")

	var cont []bool
	for _, t := range tokens {
		cont = append(cont, strings.HasPrefix(strings.ToLower(t), strings.ToLower(token)))
	}

	return any(cont)
}

func like(a, b string) bool {
	// def like(a, b):
	//...     return any([has(x, b) for x in a.split(" ")]) or any([has(x, a) for x in b.split(" ")])
	aTok := strings.Split(a, " ")
	bTok := strings.Split(b, " ")

	var aCont []bool
	var bCont []bool

	for _, t := range aTok {
		aCont = append(aCont, has(t, b))
	}

	for _, t := range bTok {
		bCont = append(bCont, has(t, a))
	}

	return any(aCont) || any(bCont)

}

// LoadCustomEngine loads a custom gval.Language to extend the capabilities of the Filters.
func LoadCustomEngine(store *Store) gval.Language {

	getProps := func(profileId string) []map[string]string {
		fmt.Println("getProps", profileId)
        result, err := store.Peers.Get(profileId)

		if err != nil {
			return nil
		}

		peer := &models.Peer{}
		_ = result.Export(peer)
		jb, err := json.Marshal(peer.RawMap["customFields"])
		fields := []map[string]string{}
		json.Unmarshal(jb, &fields)

		return fields
	}

	locMap := LoadLocationMap()
	language := gval.Full(
		gval.Function("contains", func(fullstr string, substr string) bool {
			return strings.Contains(fullstr, substr)
		}),
		gval.Function("containsInArr", func(arr []interface{}, search string) bool {
			for _, val := range arr {
				if val.(string) == search {
					return true
				}
			}
			return false
		}),
		gval.Function("zipWithin", func(sourceZip string, sourceCountry string, targetZip string, targetCountry string, distanceMeters float64) bool {
			if sourceCountry == "" || sourceZip == "" || targetZip == "" || targetCountry == "" {
				return false
			}

			source := locMap[sourceCountry][sourceZip]
			target := locMap[targetCountry][targetZip]

			return location.Distance(source[0], source[1], target[0], target[1]) <= distanceMeters
		}),
		gval.Function("coordsWithin", func(sourceLat float64, sourceLng float64, targetZip string, targetCountry string, distanceMeters float64) bool {
			target := locMap[targetCountry][targetZip]
			if targetZip == "" {
				return false
			}
			return location.Distance(sourceLat, sourceLng, target[0], target[1]) <= distanceMeters
		}),
		gval.Function("geoWithin", func(sourceLat, sourceLng, targetLat, targetLng string, distanceMeters float64) bool {
			if sourceLat == "" || sourceLng == "" || targetLat == "" || targetLng == "" {
				return false
			}
			sourceLat64, _ := strconv.ParseFloat(sourceLat, 64)
			sourceLng64, _ := strconv.ParseFloat(sourceLng, 64)
			targetLat64, _ := strconv.ParseFloat(targetLat, 64)
			targetLng64, _ := strconv.ParseFloat(targetLng, 64)
			return location.Distance(sourceLat64, sourceLng64, targetLat64, targetLng64) <= distanceMeters
		}),
		gval.Function("compareString", func(x, y string) bool {
			return x < y
		}),
		gval.Function("like", func(x, y string) bool {
			if y == "" || x == "" {
				return false
			}
			return like(x, y)
		}),

		// `getProfile(doc.peerId)["age"]["min"] > 14`
		gval.Function("getProfile", func(profileId string) map[string]interface{} {
            profile := store.Peers.Search(profileId)
			if profile.Count == 0 {
				return make(map[string]interface{})
			}
			return profile.Documents[0].ExportI()
		}),

        gval.Function("hasProp", func(profileId, target string) bool {
            fields := getProps(profileId)
            for _, field := range fields {
                if prop, ok := field["label"]; ok && prop == target {
                    return true
                }
            }
            return false
        }),

        gval.Function("getPropAsString", func(profileId string, target string) string {
			fields := getProps(profileId)
			for _, field := range fields {
				if prop, ok := field["label"]; ok && prop == target {
					return field["value"]
				}
			}
			return ""
		}),

        gval.Function("asInt", func(s string) int {
            a, _ := strconv.Atoi(s)
            return a
        }),

        gval.Function("asFloat", func(s string) float64 {
            f, _ := strconv.ParseFloat(s, 64)
            return f
        }),

	)
	return language
}
