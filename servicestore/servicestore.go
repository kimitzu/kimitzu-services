package servicestore

import (
	"fmt"
	"path"

	gomenasai "gitlab.com/nokusukun/go-menasai/manager"

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
		Name: "listings",
		Path: listingStorePath,
		IndexPaths: []string{
			"$.item.description",
			"$.item.title",
			"$.metadata.serviceClassification",
			"$.hash",
			"$.vendorID",
		},
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
