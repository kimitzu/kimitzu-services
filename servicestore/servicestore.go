package servicestore

import (
	"fmt"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
	"gitlab.com/nokusukun/go-menasai/chunk"
)

type MainStorage struct {
	PeerData map[string]*models.Peer
	Listings []*models.Listing
}

type MainManagedStorage struct {
	Pmap     map[string]string
	PeerData *chunk.Chunk
	Listings *chunk.Chunk
}

func InitializeManagedStorage() *MainManagedStorage {
	store := MainManagedStorage{}
	store.Pmap = make(map[string]string)
	peerConfig := &chunk.Config{
		ID:         "peers",
		Path:       "data/peers.chk",
		IndexDir:   "data/index_peers",
		IndexPaths: []string{"$.shortDescription"},
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

	return &store
}

// InitializeStore - Initializes and returns a MainStorage instance,
// 		pass this around the various services, acts as like the centraliezd
// 		storage for the listings and Peer Data
func InitializeStore() *MainStorage {
	store := MainStorage{}
	store.PeerData = make(map[string]*models.Peer)
	store.Listings = []*models.Listing{}
	return &store
}
