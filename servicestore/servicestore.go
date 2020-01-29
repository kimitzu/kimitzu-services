package servicestore

import (
	"fmt"
	"path"
    "sync"

	gomenasai "github.com/nokusukun/go-menasai/manager"

	"github.com/kimitzu/kimitzu-services/models"
)

// _MainStorage is defunct, user Store
type _MainStorage struct {
	PeerData map[string]*models.Peer
	Listings []*models.Listing
}

// Store holds the storage stuff
//	PMap is the peer mapping of the peerID to the chunk peerDocumentID
type Store struct {
    PMap      map[string]string
    PMapLock  *sync.RWMutex
	Peers     *gomenasai.Gomenasai
	Listings  *gomenasai.Gomenasai
	StorePath string
}

func (m *Store) SafePMapModify(function func()) {
    m.PMapLock.Lock()
    function()
    m.PMapLock.Unlock()
}

func (m *Store) PMapSet(id, val string) {
    m.SafePMapModify(func() {
        m.PMap[id] = val
    })
}

// InitializeManagedStorage - Initializes and returns a _MainStorage instance,
// 		pass this around the various services, acts as like the centraliezd
// 		storage for the listings and Peer Data
func InitializeManagedStorage(rootPath string) *Store {
	store := Store{}
    store.PMapLock = &sync.RWMutex{}
    store.PMap = make(map[string]string)
	store.StorePath = rootPath

	peerStorePath := path.Join(rootPath, "data", "peers")
	listingStorePath := path.Join(rootPath, "data", "listings")

	peerStoreConfig := &gomenasai.GomenasaiConfig{
		Name:       "peers",
		Path:       peerStorePath,
		IndexPaths: []string{"$.name", "$.shortDescription"},
	}

	listingStoreConfig := &gomenasai.GomenasaiConfig{
		Name: "listings",
		Path: listingStorePath,
		IndexPaths: []string{
			"$.item.description",
			"$.item.title",
			"$.metadata.serviceClassification",
			"$.hash",
			"$.vendorID.peerID",
		},
	}

	if gomenasai.Exists(peerStorePath) {
		peerdata, err := gomenasai.Load(peerStorePath)
		if err != nil {
			panic(fmt.Errorf("Failed to load peer database: %v", err))
		}
		store.Peers = peerdata
	} else {
		peerdata, err := gomenasai.New(peerStoreConfig)
		if err != nil {
			panic(fmt.Errorf("Failed to create listing database: %v", err))
		}
		store.Peers = peerdata
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

    store.Listings.OverrideEvalEngine(LoadCustomEngine(&store))

	return &store
}

