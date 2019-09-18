package servicestore

import (
	"strconv"
	"strings"

	"github.com/PaesslerAG/gval"

	"github.com/djali-foundation/djali-services/location"
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
func LoadCustomEngine() gval.Language {
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
			source := locMap[sourceCountry][sourceZip]
			target := locMap[targetCountry][targetZip]
			if targetZip == "" {
				return false
			}
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
			return like(x, y)
		}),
	)
	return language
}
