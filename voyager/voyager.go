package voyager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/levigross/grequests"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
)

var (
	crawledPeers []string
	pendingPeers chan string
	peerData     map[string]*models.Peer
	listings     []*models.Listing
)

func findPeers(peerlist chan<- string, log chan<- string) {
	for {
		resp, err := grequests.Get("http://localhost:4002/ob/peers", nil)
		if err != nil {
			log <- "Can't Load OpenBazaar Peers"
		}
		listJSON := []string{}
		json.Unmarshal([]byte(resp.String()), &listJSON)
		for _, peer := range listJSON {
			peerlist <- peer
		}
		time.Sleep(time.Second * 5)
	}
}

func getPeerData(peer string, log chan<- string) (string, string, error) {
	log <- "Retrieving Peer Data: " + peer
	ro := &grequests.RequestOptions{RequestTimeout: 30 * time.Second}
	profile, err := grequests.Get("http://localhost:4002/ob/profile/"+peer+"?usecache=false", ro)
	if err != nil {
		log <- fmt.Sprintln("Can't Retrieve peer data from "+peer, err)
		return "", "", fmt.Errorf("Retrieve timeout")
	}
	listings, err := grequests.Get("http://localhost:4002/ob/listings/"+peer, ro)
	if err != nil {
		log <- fmt.Sprintln("Can't Retrive listing from peer "+peer, err)
		return "", "", fmt.Errorf("Retrieve timeout")
	}

	return profile.String(), listings.String(), nil
}

func RunVoyagerService(log chan<- string) {
	log <- "Initializing"
	pendingPeers = make(chan string, 50)
	crawledPeers = []string{}
	peerData = make(map[string]*models.Peer)
	listings = []*models.Listing{}

	go findPeers(pendingPeers, log)
	// Digests found peers
	go func() {
		for {
			select {
			case peer := <-pendingPeers:
				if _, exists := peerData[peer]; !exists {
					log <- "Found Peer: " + peer
					peerDat, listingDat, err := getPeerData(peer, log)
					if err != nil {
						log <- fmt.Sprint("Error Retrieving Peer", err)
						break
					}
					peerJSON := make(map[string]interface{})
					peerListings := []*models.Listing{}

					json.Unmarshal([]byte(listingDat), &peerListings)
					json.Unmarshal([]byte(peerDat), &peerJSON)

					for _, listing := range peerListings {
						listing.PeerSlug = peer + ":" + listing.Slug
						listing.ParentPeer = peer
						listings = append(listings, listing)
					}

					log <- fmt.Sprint(" id  > ", peerJSON["name"])
					log <- fmt.Sprint(" len > ", strconv.Itoa(len(peerListings)))

					peerData[peer] = &models.Peer{
						ID:       peer,
						RawMap:   peerJSON,
						RawData:  peerDat,
						Listings: peerListings}
					peerObj, err := json.Marshal(peerData[peer])
					if err != nil {
						log <- "Failed loading to json " + peer
					}
					ioutil.WriteFile("data/peers/"+peer, []byte(fmt.Sprint(peerObj)), 1)
				} else {
					log <- "Skipping Peer[Exists]: " + peer
				}
			}
		}
	}()

	http.HandleFunc("/listings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jsn, err := json.Marshal(listings)
		if err == nil {
			fmt.Fprint(w, string(jsn))
		} else {
			fmt.Fprint(w, `{"error": "notFound"}`)
		}
	})

	http.HandleFunc("/peer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		qpeerid := r.URL.Query().Get("id")
		var result []*models.Peer
		for peerid, peer := range peerData {
			if qpeerid == peerid {
				result = append(result, peer)
			}
		}

		if len(result) != 0 {
			jsn, _ := json.Marshal(result)
			fmt.Fprint(w, string(jsn))
		} else {
			fmt.Fprint(w, `{"error": "notFound"}`)
		}
	})

	log <- "Serving at 0.0.0.0:8109"
	http.ListenAndServe(":8109", nil)
}
