package servicestore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/PaesslerAG/gval"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/location"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
	"gitlab.com/nokusukun/go-menasai/chunk"
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
	PeerData *chunk.Chunk
	Listings *chunk.Chunk
}

func LoadLocationMap() map[string]map[string][]float64 {
	fstream, err := ioutil.ReadFile("locationmap.json")
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
	)
	return language
}

// InitializeManagedStorage - Initializes and returns a MainStorage instance,
// 		pass this around the various services, acts as like the centraliezd
// 		storage for the listings and Peer Data
func InitializeManagedStorage() *MainManagedStorage {
	store := MainManagedStorage{}
	store.Pmap = make(map[string]string)
	peerConfig := &chunk.Config{
		ID:         "peers",
		Path:       "data/peers.chk",
		IndexDir:   "data/index_peers",
		IndexPaths: []string{"$.name", "$.shortDescription"},
	}

	peerdata, err := chunk.CreateChunk(peerConfig)
	if err != nil {
		fmt.Println("Storage Info: ", err)
		peerdata, err = chunk.LoadChunk(peerConfig.Path)
		if err != nil {
			panic(err)
		}
	}
	store.PeerData = peerdata

	listingConfig := &chunk.Config{
		ID:         "listing",
		Path:       "data/listings.chk",
		IndexDir:   "data/index_listings",
		IndexPaths: []string{"$.description", "$.title"},
	}

	listing, err := chunk.CreateChunk(listingConfig)
	if err != nil {
		fmt.Println("Storage Info: ", err)
		listing, err = chunk.LoadChunk(listingConfig.Path)
		if err != nil {
			panic(err)
		}
	}

	store.Listings = listing
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
