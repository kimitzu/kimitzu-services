package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type Location struct {
	Country string `json:"cou"`
	ZipCode string `json:"zip"`
	Address string `json:"add"`
	X       string `json:"x"`
	Y       string `json:"y"`
}

type LocationDistance struct {
	Loc  Location `json:"location"`
	Dist float64  `json:"distance"`
}

func RunLocationService() {
	fmt.Println("Reading...")
	fstream, err := ioutil.ReadFile("locations.json")
	if err != nil {
		fmt.Errorf("Failed Reading file", err)
	}
	obj := []Location{}
	json.Unmarshal(fstream, &obj)

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		zipCode := r.URL.Query().Get("zip")
		country := r.URL.Query().Get("country")
		address := r.URL.Query().Get("address")
		x := r.URL.Query().Get("x")
		y := r.URL.Query().Get("y")

		fmt.Println("Querying for: " + zipCode + country)
		var result []Location
		for _, loc := range obj {
			var matches []bool
			matches = append(matches, loc.ZipCode == zipCode || zipCode == "")
			matches = append(matches, loc.Country == country || country == "")
			matches = append(matches, strings.Contains(loc.Address, address) || address == "")
			matches = append(matches, loc.X == x || x == "")
			matches = append(matches, loc.Y == y || y == "")

			if !stringInSlice(false, matches) {
				result = append(result, loc)
			}
		}
		if len(result) != 0 {
			jsn, _ := json.Marshal(result)
			fmt.Fprint(w, string(jsn))
		} else {
			fmt.Fprint(w, `{"error": "notFound"}`)
		}

	})

	http.HandleFunc("/codesfrom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		x, _ := strconv.ParseFloat(r.URL.Query().Get("x"), 64)
		y, _ := strconv.ParseFloat(r.URL.Query().Get("y"), 64)
		within, _ := strconv.ParseFloat(r.URL.Query().Get("within"), 64)

		var result []LocationDistance
		for _, loc := range obj {
			tarx, _ := strconv.ParseFloat(loc.X, 64)
			tary, _ := strconv.ParseFloat(loc.Y, 64)
			dist := Distance(x, y, tarx, tary)
			if dist <= within {
				fmt.Println("Found Something Nearby: ", loc)
				fmt.Println("Distance between Locations: " + strconv.FormatFloat(dist, 'f', -1, 64))
				result = append(result, LocationDistance{loc, dist})
			}
		}

		if len(result) != 0 {
			jsn, _ := json.Marshal(result)
			fmt.Fprint(w, string(jsn))
		} else {
			fmt.Fprint(w, `{"error": "notFound"}`)
		}

	})

	fmt.Println("Serving")
	http.ListenAndServe(":8108", nil)
}

func stringInSlice(a bool, list []bool) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}
