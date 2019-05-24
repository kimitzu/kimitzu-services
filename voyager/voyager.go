package voyager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/levigross/grequests"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
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

func downloadFile(fileName string, log *servicelogger.LogPrinter) {
	if doesFileExist("data/images/" + fileName) {
		log.Verbose("File " + fileName + " already downloaded, skipping...")
		return
	}

	file, err := http.Get("http://localhost:4002/ipfs/" + fileName)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create("data/images/" + fileName)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(outFile, file.Body)
}

func digestPeer(peer string, log *servicelogger.LogPrinter, store *servicestore.MainManagedStorage) (*models.Peer, error) {
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
		store.Listings.Insert(listing)

		/**
		 * Removed due to double-save bug.
		 * @Von, please permanently remove to confirm.
		 */
		// store.Listings.Insert(listing, true)

		downloadFile(listing.Thumbnail.Medium, log)
		downloadFile(listing.Thumbnail.Small, log)
		downloadFile(listing.Thumbnail.Tiny, log)
	}

	log.Verbose(fmt.Sprint(" id  > ", peerJSON["name"]))
	log.Verbose(fmt.Sprint(" len > ", strconv.Itoa(len(peerListings))))
	return &models.Peer{
		ID:       peer,
		RawMap:   peerJSON,
		RawData:  peerDat,
		Listings: peerListings}, nil
}

// RunVoyagerService - Starts the voyager service. Handles the crawling of the nodes for the listings.
func RunVoyagerService(log *servicelogger.LogPrinter, store *servicestore.MainManagedStorage) {
	log.Info("Initializing")
	pendingPeers = make(chan string, 50)
	retryPeers = make(map[string]int)

	ensureDir("data/peers/.test")
	ensureDir("data/images/.test")
	go findPeers(pendingPeers, log)

	peers := store.PeerData.Search("")

	for _, doc := range peers.Documents {
		interfpeer := models.Peer{}
		doc.Export(&interfpeer)

		store.Pmap[interfpeer.ID] = doc.ID
	}

	// Digests found peers
	go func() {
		for peer := range pendingPeers {
			log.Debug("Digesting Peer: " + peer)
			if val, exists := retryPeers[peer]; exists && val >= 5 {
				continue
			}
			if _, exists := store.Pmap[peer]; !exists {
				log.Debug("Found Peer: " + peer)
				peerObj, err := digestPeer(peer, log, store)
				if err != nil {
					log.Error(err)
					store.Pmap[peer] = ""
					continue
				}
				peerObjID, _ := store.PeerData.Insert(peerObj.RawMap)
				store.Pmap[peer] = peerObjID
				go store.Listings.FlushSE()
				store.Listings.Commit()
				store.PeerData.Commit()
			} else {
				log.Debug("Skipping Peer[Exists]: " + peer)
			}
		}
		log.Error("Digesting stopped")
	}()

	http.HandleFunc("/djali/peers/listings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		result := store.Listings.Search("")
		jsn, err := result.ExportJSONArray()
		if err == nil {
			fmt.Fprint(w, jsn)
		} else {
			fmt.Fprint(w, `{"error": "notFound"}`)
		}
	})

	http.HandleFunc("/djali/peer/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		qpeerid := r.URL.Query().Get("id")
		docid, exists := store.Pmap[qpeerid]
		toret := ""

		if exists {
			doc, err := store.PeerData.Get(docid)
			if err != nil {
				toret = fmt.Sprintf(`{"error": "failedToRetrievePeer", "details": "%v"}`, err)
			}
			toret = string(doc.Content)
		} else {
			toret = `{"error": "notFound"}`
		}
		fmt.Fprint(w, toret)
	})

	http.HandleFunc("/djali/peer/add", func(w http.ResponseWriter, r *http.Request) {
		peerID := r.URL.Query().Get("id")

		peerObj, err := digestPeer(peerID, log, store)
		if err != nil {
			fmt.Fprint(w, `{"response": "Error adding peer to queue"}`)
		}
		for _, listing := range peerObj.Listings {
			store.Listings.Insert(listing)
		}
		peerObjID, _ := store.PeerData.Insert(peerObj.RawMap)
		store.Pmap[peerID] = peerObjID
		go store.Listings.FlushSE()
		store.Listings.Commit()
		store.PeerData.Commit()

		message := "Peer ID " + peerID + " manually added to voyager queue"
		log.Debug(message)
		fmt.Fprint(w, message)
	})

	// Deprecation Notice
	//		Please remove this snippet down the lone
	// http.HandleFunc("/djali/search", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	query := r.URL.Query().Get("query")
	// 	filter := r.URL.Query().Get("filter")
	// 	log.Verbose("[/search] Parameter [query=" + query + "]")

	// 	results := store.Listings.Search(query)
	// 	if filter != "" {
	// 		results.Filter(filter)
	// 	}

	// 	fmt.Println("Result: ", results)
	// 	resultsResponse, _ := results.ExportJSONArray()
	// 	fmt.Fprint(w, string(resultsResponse))
	// })

	http.HandleFunc("/djali/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		defer r.Body.Close()
		params := &models.AdvancedSearchQuery{}
		json.NewDecoder(r.Body).Decode(params)

		log.Verbose("[/search] Parameter [query=" + params.Query + "]")

		results := store.Listings.Search(params.Query)

		if len(params.Filters) != 0 {
			for _, filter := range params.Filters {
				log.Debug("Running filter: " + filter)
				results.Filter(filter)
			}
		}

		if params.Limit != 0 {
			results.Limit(params.Start, params.Limit)
		}

		resultsResponse, _ := results.ExportJSONArray()
		fmt.Fprint(w, string(resultsResponse))
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

func doesFileExist(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
