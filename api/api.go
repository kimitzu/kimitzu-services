package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicelogger"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicestore"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/voyager"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
)

func RunHTTPService(log *servicelogger.LogPrinter, store *servicestore.MainManagedStorage) {
	log.Info("Starting HTTP Service")

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

	http.HandleFunc("/djali/peers", func(w http.ResponseWriter, r *http.Request) {
		peers := store.PeerData.Search("")
		data, _ := peers.ExportJSONArray()
		fmt.Fprint(w, string(data))
	})

	http.HandleFunc("/djali/peer/add", func(w http.ResponseWriter, r *http.Request) {
		peerID := r.URL.Query().Get("id")

		peerObj, err := voyager.DigestPeer(peerID, store)
		if err != nil {
			fmt.Fprint(w, `{"response": "Error adding peer to queue"}`)
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

	http.HandleFunc("/djali/peer/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		params := &models.AdvancedSearchQuery{}
		err = json.Unmarshal(b, &params)
		if err != nil {
			fmt.Fprint(w, fmt.Sprintf(`{"error": "Failed to decode body", "goerror": "%v"}`, err))
			return
		}

		log.Verbose("[/peer/search] Parameter [query=" + params.Query + "]")

		results := store.PeerData.Search(params.Query)

		if len(params.Filters) != 0 {
			for _, filter := range params.Filters {
				log.Debug("Running filter: " + filter)
				results.Filter(filter)
			}
		}

		if params.Limit != 0 {
			results.Limit(params.Start, params.Limit)
		}

		if params.Sort != "" {
			results.Sort(params.Sort)
		}

		if len(params.Transforms) != 0 {
			d, _ := json.Marshal(params.Transforms)
			results.Transform(string(d))
		}

		resultsResponse, _ := results.ExportJSONArray()
		fmt.Fprint(w, string(resultsResponse))
	})

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
	http.HandleFunc("/djali/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		params := &models.AdvancedSearchQuery{}
		err = json.Unmarshal(b, &params)
		if err != nil {
			fmt.Fprint(w, fmt.Sprintf(`{"error": "Failed to decode body", "goerror": "%v"}`, err))
			return
		}

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

		if params.Sort != "" {
			results.Sort(params.Sort)
		}

		if len(params.Transforms) != 0 {
			d, _ := json.Marshal(params.Transforms)
			results.Transform(string(d))
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
