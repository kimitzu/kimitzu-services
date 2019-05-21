package voyager

import (
	"encoding/json"
	"fmt"
	"io"
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
			log.Error("Can't Load OpenBazaar Peers")
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

func downloadFile(fileName string) error {
	file, err := http.Get("http://localhost:4002/ipfs/" + fileName)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create("data/images/" + fileName)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(outFile, file.Body)
	return err
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

		downloadFile(listing.Thumbnail.Medium)
		downloadFile(listing.Thumbnail.Small)
		downloadFile(listing.Thumbnail.Tiny)
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
			continue
		}

		peerInfo := models.Peer{}
		json.Unmarshal(peer, &peerInfo)

		store.Listings = append(store.Listings, peerInfo.Listings...)
		// for _, listing := range peerInfo.Listings {
		// 	store.Listings = append(store.Listings, listing)
		// }

		store.PeerData[peerInfo.ID] = &peerInfo
	}
}

// RunVoyagerService - Starts the voyager service. Handles the crawling of the nodes for the listings.
func RunVoyagerService(log *servicelogger.LogPrinter, store *servicestore.MainStorage) {
	log.Info("Initializing")
	pendingPeers = make(chan string, 50)
	crawledPeers = []string{}
	retryPeers = make(map[string]int)
	// peerData = make(map[string]*models.Peer)
	// listings = []*models.Listing{}

	ensureDir("data/peers/.test")
	ensureDir("data/images/.test")
	go findPeers(pendingPeers, log)

	Initialize(log, store)
	queryEngine := search.InitializeQueryEngine(log, 2)

	// Digests found peers
	go func() {
		for peer := range pendingPeers {
			log.Debug("Digesting Peer: " + peer)
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
		// var result []*models.Peer
		// for peerid, peer := range store.PeerData {
		// 	if qpeerid == peerid {
		// 		result = append(result, peer)
		// 	}
		// }
		result, exists := store.PeerData[qpeerid]

		if exists {
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
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

	http.HandleFunc("/advquery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		defer r.Body.Close()
		params := &search.QueryParameters{}
		json.NewDecoder(r.Body).Decode(params)
		params.Collection = store.Listings
		params.WorkerCount = 2
		results := queryEngine.QueryListings(params)
		if results != nil {
			resultsResponse, _ := json.Marshal(results)
			fmt.Fprint(w, string(resultsResponse))
		} else {
			fmt.Fprint(w, `{"error": "No more documents to return."}`)
		}
	})

	http.HandleFunc("/djali/media", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		image, err := os.Open("data/images/" + id)
		if err != nil {
			fmt.Fprint(w, `{"response": "Media not found"}`)
		}

		// Setup response headers
		fileHeader := make([]byte, 512)
		image.Read(fileHeader)
		contentType := http.DetectContentType(fileHeader)
		stat, _ := image.Stat()
		size := strconv.FormatInt(stat.Size(), 10)

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", size)
		image.Seek(0, 0)
		io.Copy(w, image)
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
