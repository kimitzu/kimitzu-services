package voyager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/levigross/grequests"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/search"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicelogger"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicestore"
)

var (
	crawledPeers []string
	pendingPeers chan string
	retryPeers   map[string]int
)

func findPeers(peerlist chan<- string, log *servicelogger.LogPrinter) {
	for {
		resp, err := grequests.Get("http://localhost:4002/ob/peers", nil)
		if err != nil {
			log.Info("Can't Load OpenBazaar Peers")
		}
		listJSON := []string{}
		json.Unmarshal([]byte(resp.String()), &listJSON)
		for _, peer := range listJSON {
			peerlist <- peer
		}
		time.Sleep(time.Second * 5)
	}
}

func getPeerData(peer string, log *servicelogger.LogPrinter) (string, string, error) {
	log.Debug("Retrieving Peer Data: " + peer)
	ro := &grequests.RequestOptions{RequestTimeout: 30 * time.Second}
	profile, err := grequests.Get("http://localhost:4002/ob/profile/"+peer+"?usecache=false", ro)
	if err != nil {
		log.Error(fmt.Sprintln("Can't Retrieve peer data from "+peer, err))
		return "", "", fmt.Errorf("Retrieve timeout")
	}
	listings, err := grequests.Get("http://localhost:4002/ob/listings/"+peer, ro)
	if err != nil {
		log.Error(fmt.Sprintln("Can't Retrive listing from peer "+peer, err))
		return "", "", fmt.Errorf("Retrieve timeout")
	}

	return profile.String(), listings.String(), nil
}

func digestPeer(peer string, log *servicelogger.LogPrinter, store *servicestore.MainStorage) (*models.Peer, error) {
	peerDat, listingDat, err := getPeerData(peer, log)
	if err != nil {
		val, exists := retryPeers[peer]
		if !exists {
			retryPeers[peer] = 1
		} else {
			retryPeers[peer]++
		}
		return nil, fmt.Errorf(fmt.Sprint("["+strconv.Itoa(val)+"] Error Retrieving Peer ", err))
	}
	peerJSON := make(map[string]interface{})
	peerListings := []*models.Listing{}

	json.Unmarshal([]byte(listingDat), &peerListings)
	json.Unmarshal([]byte(peerDat), &peerJSON)

	for _, listing := range peerListings {
		listing.PeerSlug = peer + ":" + listing.Slug
		listing.ParentPeer = peer
		store.Listings = append(store.Listings, listing)
	}

	log.Verbose(fmt.Sprint(" id  > ", peerJSON["name"]))
	log.Verbose(fmt.Sprint(" len > ", strconv.Itoa(len(peerListings))))
	return &models.Peer{
		ID:       peer,
		RawMap:   peerJSON,
		RawData:  peerDat,
		Listings: peerListings}, nil
}

func Initialize(log *servicelogger.LogPrinter, store *servicestore.MainStorage) {
	log.Info("Initializing precrawled listing information...")
	files, err := ioutil.ReadDir("data/peers")
	if err != nil {
		fmt.Println("Error reading data/peers directory")
	}
	for _, file := range files {
		peer, err := ioutil.ReadFile("data/peers/" + file.Name())
		if err != nil {
			fmt.Println("Error reading data/peers/" + file.Name())
		}

		peerInfo := models.Peer{}
		json.Unmarshal(peer, &peerInfo)

		for _, listing := range peerInfo.Listings {
			store.Listings = append(store.Listings, listing)
		}

		store.PeerData[peerInfo.ID] = &peerInfo
	}
}

func RunVoyagerService(log *servicelogger.LogPrinter, store *servicestore.MainStorage) {
	log.Info("Initializing")
	pendingPeers = make(chan string, 50)
	crawledPeers = []string{}
	retryPeers = make(map[string]int)
	// peerData = make(map[string]*models.Peer)
	// listings = []*models.Listing{}

	ensureDir("data/peers/.test")
	go findPeers(pendingPeers, log)

	Initialize(log, store)

	// Digests found peers
	go func() {
		for {
			select {
			case peer := <-pendingPeers:
				if val, exists := retryPeers[peer]; exists && val >= 5 {
					break
				}
				if _, exists := store.PeerData[peer]; !exists {
					log.Debug("Found Peer: " + peer)
					peerObj, err := digestPeer(peer, log, store)
					if err != nil {
						log.Error(err)
						break
					}
					store.PeerData[peer] = peerObj
					peerStr, err := json.Marshal(store.PeerData[peer])
					if err != nil {
						log.Error("Failed loading to json " + peer)
					}

					ioutil.WriteFile("data/peers/"+peer, peerStr, 1)
				} else {
					log.Debug("Skipping Peer[Exists]: " + peer)
				}
			}
		}
	}()

	http.HandleFunc("/djali/peers/listings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jsn, err := json.Marshal(store.Listings)
		if err == nil {
			fmt.Fprint(w, string(jsn))
		} else {
			fmt.Fprint(w, `{"error": "notFound"}`)
		}
	})

	http.HandleFunc("/djali/peer/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		qpeerid := r.URL.Query().Get("id")
		var result []*models.Peer
		for peerid, peer := range store.PeerData {
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

	http.HandleFunc("/djali/peer/add", func(w http.ResponseWriter, r *http.Request) {
		peerID := r.URL.Query().Get("id")
		digestPeer(peerID, log, store)
		message := "Peer ID " + peerID + " manually added to voyager queue"
		log.Debug(message)
		fmt.Fprint(w, message)
	})

	http.HandleFunc("/djali/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("query")
		averageRating, err := strconv.ParseInt(r.URL.Query().Get("averageRating"), 10, 64)
		if err != nil {
			log.Error("Conversion error in /search/?averageRating")
		}
		log.Verbose("[/search] Parameter [query=" + query + "]")
		results := search.Find(query, averageRating, store.Listings)
		resultsResponse, _ := json.Marshal(results)
		fmt.Fprint(w, string(resultsResponse))
	})

	log.Info("Serving at 0.0.0.0:8109")
	http.ListenAndServe(":8109", nil)
}

func ensureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
}
