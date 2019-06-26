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
	peerStream   chan string
	retryPeers   map[string]int
	log          *servicelogger.LogPrinter
)

var ro = &grequests.RequestOptions{RequestTimeout: 70 * time.Second}
var maxClosest = make(chan int, 5)

func findClosestPeers(peer string, peerlist chan<- string) {
	// This makes sure that the findClosestPeers doesn't overfill the requests
	// by limiting it to 5 concurrent calls.
	maxClosest <- 1
	defer func() {
		<-maxClosest
	}()
	log.Debug(fmt.Sprintf("Retrieving closest peers for %v", peer))
	resp, err := grequests.Get("http://localhost:4002/ob/closestpeers/"+peer, ro)
	if err != nil {
		log.Error("Peer resolve timeout for " + peer)
	}
	listJSON := []string{}
	json.Unmarshal([]byte(resp.String()), &listJSON)
	for _, peer := range listJSON {
		peerlist <- peer
	}
}

func findPeers(peerlist chan<- string) {
	for {
		resp, err := grequests.Get("http://localhost:4002/ob/peers", ro)
		if err != nil {
			log.Error("Can't Load OpenBazaar Peers")
			continue
		}
		listJSON := []string{}
		json.Unmarshal([]byte(resp.String()), &listJSON)
		for _, peer := range listJSON {
			peerlist <- peer
			go findClosestPeers(peer, peerlist)
		}
		time.Sleep(time.Second * 5)
	}
}

func getPeerData(peer string) (string, string, error) {
	log.Debug("Retrieving Peer Data: " + peer)

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

func downloadFile(fileName string) {
	if doesFileExist("data/images/" + fileName) {
		// log.Verbose("File " + fileName + " already downloaded, skipping...")
		return
	}

	file, err := http.Get("http://localhost:4002/ipfs/" + fileName)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create("data/images/" + fileName)
	defer outFile.Close()
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(outFile, file.Body)
}

// DigestPeer downloads the peer data and package it in a easy to use struct.
//		Downloads the listings and stores them in the database as well.
func DigestPeer(peer string, store *servicestore.MainManagedStorage) (*models.Peer, error) {
	peerDat, listingDat, err := getPeerData(peer)
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

	if peerJSON["success"] != nil {
		if !peerJSON["success"].(bool) {
			return nil, fmt.Errorf(peerJSON["reason"].(string))
		}
	}

	for _, listing := range peerListings {
		listing.PeerSlug = peer + ":" + listing.Slug
		listing.ParentPeer = peer
		ro := &grequests.RequestOptions{RequestTimeout: 30 * time.Second}
		listingData, err := grequests.Get("http://localhost:4002/ob/listing/"+peer+"/"+listing.Slug, ro)

		if err != nil {
			log.Verbose(fmt.Sprintf("Failed to retrieve IPFS data of %v", listing.PeerSlug))
			continue
		}
		ipfsListing := models.IPFSListing{}

		json.Unmarshal([]byte(listingData.String()), &ipfsListing)

		// Shuffle the old listing model into the newer listing model
		// by unmarshalling the data from the old model to the new one
		// this is because the /ob/listing data needs to coalesce with
		// the old model. It's hacky I know, but GO doesn't really have
		// an equivalent to Python's dict.update()
		classListing := ipfsListing.Listing
		oldListingDat, err := json.Marshal(listing)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(oldListingDat, &classListing)

		// Check if the listing hash already exists and update it instead of inserting a new one.
		existing := store.Listings.Search(classListing.Hash)
		if existing.Count == 1 {
			store.Listings.Update(existing.Documents[0].ID, classListing)
		} else {
			store.Listings.Insert(classListing)
		}

		downloadFile(listing.Thumbnail.Medium)
		downloadFile(listing.Thumbnail.Small)
		downloadFile(listing.Thumbnail.Tiny)
	}

	log.Verbose(fmt.Sprint(" id  > ", peerJSON["name"]))
	log.Verbose(fmt.Sprint(" len > ", strconv.Itoa(len(peerListings))))
	return &models.Peer{
		ID:       peer,
		RawMap:   peerJSON,
		LastPing: time.Now().Unix()}, nil
}

func DigestService(peerStream chan string, store *servicestore.MainManagedStorage) {
	for peer := range peerStream {
		log.Debug("Recieved peer...")
		if val, exists := retryPeers[peer]; exists && val >= 5 {
			continue
		}
		if _, exists := store.Pmap[peer]; !exists {
			log.Debug("Digesting Peer: " + peer)
			log.Debug("Found Peer: " + peer)
			peerObj, err := DigestPeer(peer, store)
			if err != nil {
				log.Error(err)
				store.Pmap[peer] = ""
				continue
			}
			peerObjID, err := store.PeerData.Insert(peerObj)
			if err != nil {
				panic(err)
			}
			store.Pmap[peer] = peerObjID
			go store.Listings.FlushSE()
			store.Listings.Commit()
			store.PeerData.Commit()
		} else {
			log.Debug("Peer alreaday exists: " + peer)
		}
		log.Debug("Getting peer from peerStream...")
	}
	log.Error("Digesting stopped")
}

func IsPeerOnline(peerid string) bool {
	isOnline, err := grequests.Get("http://localhost:4002/ob/peerinfo/"+peerid+"?usecache=false", ro)
	if err != nil {
		return false
	}
	result := make(map[string]string)
	isOnline.JSON(&result)
	return result["result"] == "online"
}

// RunVoyagerService - Starts the voyager service. Handles the crawling of the nodes for the listings.
func RunVoyagerService(logP *servicelogger.LogPrinter, store *servicestore.MainManagedStorage) {
	log = logP
	log.Info("Starting Voyager Service")
	peerStream = make(chan string, 1000)
	retryPeers = make(map[string]int)

	ensureDir("data/peers/.test")
	ensureDir("data/images/.test")
	go findPeers(peerStream)

	peers := store.PeerData.Search("")

	for _, doc := range peers.Documents {
		interfpeer := models.Peer{}
		doc.Export(&interfpeer)

		store.Pmap[interfpeer.ID] = doc.ID
	}

	// Digests found peers
	go DigestService(peerStream, store)

	// Occasionally ping the peers
	go func() {
		for {
			peers := store.PeerData.Search("")
			for _, peerd := range peers.Documents {
				peer := models.Peer{}
				peerd.Export(&peer)
				log.Verbose(fmt.Sprintf("Pinging %v", peer.ID))
				if IsPeerOnline(peer.ID) {
					peer.LastPing = time.Now().Unix()
					store.PeerData.Update(peerd.ID, peer)
					DigestPeer(peer.ID, store)
				}
			}
			time.Sleep(time.Minute * 30)
		}
	}()

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
