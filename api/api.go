package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/djali-foundation/djali-services/location"

	"github.com/djali-foundation/djali-services/servicelogger"
	"github.com/djali-foundation/djali-services/servicestore"
	"github.com/djali-foundation/djali-services/voyager"

	"github.com/djali-foundation/djali-services/models"
)

var (
	store *servicestore.MainManagedStorage
)

type APIListResult struct {
	Count     int           `json:"count"`
	Limit     int           `json:"limit"`
	NextStart int           `json:"nextStart"`
	Data      []interface{} `json:"data"`
}

func setupResponse(w *http.ResponseWriter, req *http.Request) bool {
	(*w).Header().Set("Access-Control-Allow-Origin", req.Header.Get("origin"))
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, PATCH, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Origin, X-Requested-With")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Content-Type", "application/json")
	if req.Method == "OPTIONS" {
		(*w).WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func HTTPFlushAll(w http.ResponseWriter, r *http.Request) {
	store.Listings.FlushSE()
	store.PeerData.FlushSE()
	fmt.Fprint(w, `{"result": "ok"}`)
}

func HTTPPeerGetListings(w http.ResponseWriter, r *http.Request) {
	if retOK := setupResponse(&w, r); retOK {
		return
	}

	result := store.Listings.Search("")
	jsn, err := result.ExportJSONArray()
	if err == nil {
		fmt.Fprint(w, jsn)
	} else {
		fmt.Fprint(w, `{"error": "notFound"}`)
	}
}

func HTTPPeerGet(w http.ResponseWriter, r *http.Request) {
	if retOK := setupResponse(&w, r); retOK {
		return
	}

	qpeerid := r.URL.Query().Get("id")
	force := r.URL.Query().Get("force")
	docid, exists := store.Pmap[qpeerid]
	toret := ""
	message := ""

	if exists && docid != "" && force != "true" {
		doc, err := store.PeerData.Get(docid)
		if err != nil {
			toret = fmt.Sprintf(`{"error": "failedToRetrievePeer", "details": "%v"}`, err)
		} else {
			toret = string(doc.Content)
		}
	} else {
		//log.Error("Peer not found or forced, attempting to digest...")

		peerObj, err := voyager.DigestPeer(qpeerid, store)
		if err != nil {
			//log.Error(err)
			store.Pmap[qpeerid] = ""
			message = "failed"
		}
		peerObjID, err := store.PeerData.Insert(peerObj)
		if err != nil {
			//log.Error(err)
			message = "failed"
			toret = `{"error": "` + message + `"}`
		}

		if message != "failed" {
			store.Pmap[qpeerid] = peerObjID
			go store.Listings.FlushSE()
			store.Listings.Commit()
			store.PeerData.Commit()
			peerObjJSON, err := json.Marshal(peerObj)
			if err != nil {
				toret = `{"error": "` + err.Error() + `"}`
			}
			toret = string(peerObjJSON)
		} else {
			toret = `{"error": "Not found and failed to digest"}`
		}

	}
	if strings.Contains(toret, "error") {
		http.Error(w, toret, 500)
	} else {
		fmt.Fprint(w, toret)
	}
}

func HTTPPeers(w http.ResponseWriter, r *http.Request) {
	if retOK := setupResponse(&w, r); retOK {
		return
	}

	peers := store.PeerData.Search("")
	data, _ := peers.ExportJSONArray()
	fmt.Fprint(w, string(data))
}

func HTTPPeerAdd(w http.ResponseWriter, r *http.Request) {
	if retOK := setupResponse(&w, r); retOK {
		return
	}

	peerID := r.URL.Query().Get("id")

	message := "success"
	peerObj, err := voyager.DigestPeer(peerID, store)
	if err != nil {
		// log.Error(err)
		store.Pmap[peerID] = ""
		message = "failed"
	}
	peerObjID, err := store.PeerData.Insert(peerObj)
	if err != nil {
		// panic(err)
		message = "failed"
	}

	if message != "failed" {
		store.Pmap[peerID] = peerObjID
		go store.Listings.FlushSE()
		store.Listings.Commit()
		store.PeerData.Commit()
	}

	fmt.Fprint(w, "{\"result\": \""+message+"\"}")
}

func HTTPPeerSearch(w http.ResponseWriter, r *http.Request) {
	if retOK := setupResponse(&w, r); retOK {
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	params := &models.AdvancedSearchQuery{}
	err = json.Unmarshal(b, &params)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to decode body", "goerror": "%v"}`, err), 500)
		return
	}

	//log.Verbose("[/peer/search] Parameter [query=" + params.Query + "]")

	results := store.PeerData.Search(params.Query)

	if len(params.Filters) != 0 {
		for _, filter := range params.Filters {
			// log.Debug("Running filter: " + filter)
			results.Filter(filter)
		}
	}

	if params.Sort != "" {
		results.Sort(params.Sort)
	}

	if params.Limit != 0 {
		results.Limit(params.Start, params.Limit)
	}

	if len(params.Transforms) != 0 {
		d, _ := json.Marshal(params.Transforms)
		results.Transform(string(d))
	}

	nextStart := params.Start + params.Limit
	if nextStart >= results.Count {
		nextStart = -1
	}

	arr := []interface{}{}
	for _, doc := range results.Documents {
		i := new(interface{})
		json.Unmarshal(doc.Content, &i)
		arr = append(arr, i)
	}

	listreturn := APIListResult{
		Count:     results.Count,
		Limit:     params.Limit,
		NextStart: nextStart,
		Data:      arr,
	}
	retStr, _ := json.Marshal(listreturn)
	fmt.Fprint(w, string(retStr))
}

func HTTPListing(w http.ResponseWriter, r *http.Request) {
	if retOK := setupResponse(&w, r); retOK {
		return
	}

	hash := r.URL.Query().Get("hash")
	results := store.Listings.Search(hash)
	if len(results.Documents) == 0 {
		http.Error(w, "{\"error\": \"No results\"}", 404)
		return
	}
	listing, _ := json.Marshal(results.Documents[0].ExportI())
	fmt.Fprintf(w, string(listing))
	return
}

// Kazaam Specs https://github.com/qntfy/kazaam
/**
* {
* 	"query": "comics",
* 	"filters": [
* 		"contains(doc.slug, \"golden\")"	// Gval expression
* 	],
* 	"limit": 5,
* 	"transforms": [{						// Kazaam Spec
* 		"operation": "shift",
* 		"spec": {
* 		  "title": "title",
* 		  "owner": "parentPeer",
* 		  "price": "price.amount",
* 		  "thumb": "thumbnail.tiny"
* 		}
* 	}]
* 	}
 */
func HTTPListingSearch(w http.ResponseWriter, r *http.Request) {
	if retOK := setupResponse(&w, r); retOK {
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	params := &models.AdvancedSearchQuery{}
	err = json.Unmarshal(b, &params)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to decode body", "goerror": "%v"}`, err), 500)
		return
	}

	//log.Verbose("[/search] Parameter [query=" + params.Query + "]")

	results := store.Listings.Search(params.Query)

	if len(params.Filters) != 0 {
		for _, filter := range params.Filters {
			//log.Debug("Running filter: " + filter)
			results.Filter(filter)
		}
	}

	if params.Sort != "" {
		results.Sort(params.Sort)
	}

	if params.Limit != 0 {
		results.Limit(params.Start, params.Limit)
	}

	if len(params.Transforms) != 0 {
		d, _ := json.Marshal(params.Transforms)
		results.Transform(string(d))
	}

	nextStart := params.Start + params.Limit
	if nextStart >= results.Count {
		nextStart = -1
	}

	arr := []interface{}{}
	for _, doc := range results.Documents {
		i := new(interface{})
		json.Unmarshal(doc.Content, &i)
		arr = append(arr, i)
	}

	listreturn := APIListResult{
		Count:     results.Count,
		Limit:     params.Limit,
		NextStart: nextStart,
		Data:      arr,
	}
	retStr, _ := json.Marshal(listreturn)
	fmt.Fprint(w, string(retStr))
}

func HTTPMedia(w http.ResponseWriter, r *http.Request) {
	if retOK := setupResponse(&w, r); retOK {
		return
	}

	id := r.URL.Query().Get("id")
	image, err := os.Open("data/images/" + id)

	if err != nil {
		// If media is not found in data/images, fallback to ipfs
		resp, err2 := http.Get("http://localhost:4002/ob/images/" + id)
		if err2 != nil {
			http.Error(w, `{"error": "Media not found"}`, 404)
			return
		}
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		io.Copy(w, resp.Body)
		resp.Body.Close()
		return
	}

	fileResponder(image, w)
}

func AppendAPIService(mux *http.ServeMux) {
	mux.HandleFunc("/djali/location/query", location.HTTPLocationQueryHandler)
	mux.HandleFunc("/djali/location/codesfrom", location.HTTPLocationCodesfromHandler)

	mux.HandleFunc("/djali/peers/listings", HTTPPeerGetListings)
	mux.HandleFunc("/djali/peer/get", HTTPPeerGet)
	mux.HandleFunc("/djali/peers", HTTPPeers)
	mux.HandleFunc("/djali/peer/add", HTTPPeerAdd)
	mux.HandleFunc("/djali/peer/search", HTTPPeerSearch)

	mux.HandleFunc("/djali/listing", HTTPListing)
	mux.HandleFunc("/djali/search", HTTPListingSearch)

	mux.HandleFunc("/djali/media", HTTPMedia)
}

func AttachStore(store_ *servicestore.MainManagedStorage) {
	store = store_
}

func RunHTTPService(log *servicelogger.LogPrinter) {
	log.Info("Starting HTTP Service")

	http.HandleFunc("/djali/location/query", location.HTTPLocationQueryHandler)
	http.HandleFunc("/djali/location/codesfrom", location.HTTPLocationCodesfromHandler)

	http.HandleFunc("/djali/peers/listings", HTTPPeerGetListings)
	http.HandleFunc("/djali/peer/get", HTTPPeerGet)
	http.HandleFunc("/djali/peers", HTTPPeers)
	http.HandleFunc("/djali/peer/add", HTTPPeerAdd)
	http.HandleFunc("/djali/peer/search", HTTPPeerSearch)

	http.HandleFunc("/djali/listing", HTTPListing)
	http.HandleFunc("/djali/search", HTTPListingSearch)

	http.HandleFunc("/djali/media", HTTPMedia)

	http.HandleFunc("/authenticate", Authenticate)

	http.HandleFunc("/debug/flush", HTTPFlushAll)

	log.Info("Serving at 0.0.0.0:8109")
	http.ListenAndServe(":8109", nil)
}

func fileResponder(file *os.File, w http.ResponseWriter) {
	// Setup response headers
	fileHeader := make([]byte, 512)

	file.Read(fileHeader)
	contentType := http.DetectContentType(fileHeader)
	stat, _ := file.Stat()
	size := strconv.FormatInt(stat.Size(), 10)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", size)
	file.Seek(0, 0)
	io.Copy(w, file)
}
