package servicestore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	gomenasai "gitlab.com/nokusukun/go-menasai/manager"

	"github.com/PaesslerAG/gval"
	"github.com/djali-foundation/djali-services/location"
	"github.com/djali-foundation/djali-services/models"
)

// MainStorage is defunct, user MainManagedStorage
type MainStorage struct {
	PeerData map[string]*models.Peer
	Listings []*models.Listing
}

// MainManagedStorage holds the storage stuff
//	Pmap is the peer mapping of the peerID to the chunk peerDocumentID
type MainManagedStorage struct {
	Pmap     map[string]string
	PeerData *gomenasai.Gomenasai
	Listings *gomenasai.Gomenasai
}

func LoadLocationMap() map[string]map[string][]float64 {
	fstream, err := ioutil.ReadFile("./locationmap.json")
	if err != nil {
		fmt.Printf("Failed Reading file %v\n", err)
	}
	obj := make(map[string]map[string][]float64)
	json.Unmarshal(fstream, &obj)
	return obj
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
	)
	return language
}

// InitializeManagedStorage - Initializes and returns a MainStorage instance,
// 		pass this around the various services, acts as like the centraliezd
// 		storage for the listings and Peer Data
func InitializeManagedStorage(rootPath string) *MainManagedStorage {
	store := MainManagedStorage{}
	store.Pmap = make(map[string]string)

	peerStorePath := path.Join(rootPath, "data", "peers")
	listingStorePath := path.Join(rootPath, "data", "listings")

	peerStoreConfig := &gomenasai.GomenasaiConfig{
		Name:       "peers",
		Path:       peerStorePath,
		IndexPaths: []string{"$.profile.name", "$.profile.shortDescription"},
	}

	listingStoreConfig := &gomenasai.GomenasaiConfig{
		Name:       "listings",
		Path:       listingStorePath,
		IndexPaths: []string{"$.item.description", "$.item.title", "$.hash", "$.vendorID"},
	}

	if gomenasai.Exists(peerStorePath) {
		peerdata, err := gomenasai.Load(peerStorePath)
		if err != nil {
			panic(fmt.Errorf("Failed to load peer database: %v", err))
		}
		store.PeerData = peerdata
	} else {
		peerdata, err := gomenasai.New(peerStoreConfig)
		if err != nil {
			panic(fmt.Errorf("Failed to create listing database: %v", err))
		}
		store.PeerData = peerdata
	}

	if gomenasai.Exists(listingStorePath) {
		listing, err := gomenasai.Load(listingStorePath)
		if err != nil {
			panic(fmt.Errorf("Failed to load listing database: %v", err))
		}
		store.Listings = listing
	} else {
		listing, err := gomenasai.New(listingStoreConfig)
		if err != nil {
			panic(fmt.Errorf("Failed to create listing databse: %v", err))
		}
		store.Listings = listing
	}

	store.Listings.OverrideEvalEngine(LoadCustomEngine())

	return &store
}

// InitializeStore -  is defunct, use InitializeManagedStorage
// InitializeStore - Initializes and returns a MainStorage instance,
// 		pass this around the various services, acts as like the centraliezd
// 		storage for the listings and Peer Data
func InitializeStore() *MainStorage {
	store := MainStorage{}
	store.PeerData = make(map[string]*models.Peer)
	store.Listings = []*models.Listing{}
	return &store
}
