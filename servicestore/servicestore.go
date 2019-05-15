package servicestore

import (
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
)

type MainStorage struct {
	PeerData map[string]*models.Peer
	Listings []*models.Listing
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
