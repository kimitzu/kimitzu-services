package location

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/djali-foundation/djali-services/servicelogger"
)

var (
	obj []Location
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

func InitializeLocationService(log *servicelogger.LogPrinter) []Location {
	log.Info("Initializing")
	fstream, err := ioutil.ReadFile("location_data.json")
	if err != nil {
		fmt.Errorf("Failed Reading file %s", err)
	}
	obj := []Location{}
	json.Unmarshal(fstream, &obj)
	return obj
}

func HTTPLocationCodesfromHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	x, _ := strconv.ParseFloat(r.URL.Query().Get("x"), 64)
	y, _ := strconv.ParseFloat(r.URL.Query().Get("y"), 64)
	within, _ := strconv.ParseFloat(r.URL.Query().Get("within"), 64)

	result := getNearbyLocations(x, y, within, obj)

	if len(result) != 0 {
		jsn, _ := json.Marshal(result)
		fmt.Fprint(w, string(jsn))
	} else {
		fmt.Fprint(w, `{"error": "notFound"}`)
	}

}

func HTTPLocationQueryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	zipCode := r.URL.Query().Get("zip")
	country := r.URL.Query().Get("country")
	address := r.URL.Query().Get("address")
	x := r.URL.Query().Get("x")
	y := r.URL.Query().Get("y")

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

}

func RunLocationService(log *servicelogger.LogPrinter) {
	obj = InitializeLocationService(log)

	// http.HandleFunc("/djali/location/query", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	zipCode := r.URL.Query().Get("zip")
	// 	country := r.URL.Query().Get("country")
	// 	address := r.URL.Query().Get("address")
	// 	x := r.URL.Query().Get("x")
	// 	y := r.URL.Query().Get("y")

	// 	log.Info(fmt.Sprintln("Querying", r.URL.Query()))
	// 	var result []Location
	// 	for _, loc := range obj {
	// 		var matches []bool
	// 		matches = append(matches, loc.ZipCode == zipCode || zipCode == "")
	// 		matches = append(matches, loc.Country == country || country == "")
	// 		matches = append(matches, strings.Contains(loc.Address, address) || address == "")
	// 		matches = append(matches, loc.X == x || x == "")
	// 		matches = append(matches, loc.Y == y || y == "")

	// 		if !stringInSlice(false, matches) {
	// 			result = append(result, loc)
	// 		}
	// 	}
	// 	if len(result) != 0 {
	// 		jsn, _ := json.Marshal(result)
	// 		fmt.Fprint(w, string(jsn))
	// 	} else {
	// 		fmt.Fprint(w, `{"error": "notFound"}`)
	// 	}

	// })

	// http.HandleFunc("/djali/location/codesfrom", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	x, _ := strconv.ParseFloat(r.URL.Query().Get("x"), 64)
	// 	y, _ := strconv.ParseFloat(r.URL.Query().Get("y"), 64)
	// 	within, _ := strconv.ParseFloat(r.URL.Query().Get("within"), 64)

	// 	result := getNearbyLocations(x, y, within, obj)

	// 	if len(result) != 0 {
	// 		jsn, _ := json.Marshal(result)
	// 		fmt.Fprint(w, string(jsn))
	// 	} else {
	// 		fmt.Fprint(w, `{"error": "notFound"}`)
	// 	}

	// })

	// log.Info("Serving at 0.0.0.0:8108")
	// http.ListenAndServe(":8108", nil)
}

func getNearbyLocations(x float64, y float64, radius float64, obj []Location) []LocationDistance {
	var result []LocationDistance
	for _, loc := range obj {
		tarx, _ := strconv.ParseFloat(loc.X, 64)
		tary, _ := strconv.ParseFloat(loc.Y, 64)
		dist := Distance(x, y, tarx, tary)
		if dist <= radius {
			result = append(result, LocationDistance{loc, dist})
		}
	}
	return result
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
