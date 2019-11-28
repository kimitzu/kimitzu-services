package p2p

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "net/http/pprof"
    _ "net/http/pprof"
    "time"

    "github.com/gorilla/mux"
    "github.com/perlin-network/noise/skademlia"

    "github.com/json-iterator/go"

    "github.com/nokusukun/particles/satellite"

    "github.com/djali-foundation/djali-services/models"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type WriteRequest struct {
    PacketType  string      `json:"type"`
    Destination string      `json:"destination"`
    Namespace   string      `json:"namespace"`
    Content     interface{} `json:"content"`
}

func AttachAPI(sat *satellite.Satellite, router *mux.Router) *mux.Router {
    // router := mux.NewRouter()

    router.HandleFunc("/debug/pprof/", pprof.Index)
    router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    router.HandleFunc("/debug/pprof/profile", pprof.Profile)
    router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

    // Manually add support for paths linked to by index page at /debug/pprof/
    router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
    router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
    router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
    router.Handle("/debug/pprof/block", pprof.Handler("block"))

    router.HandleFunc("/p2p/peers", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Retieving Peers")
        var ids []string

        for id, _ := range sat.Peers {
            ids = append(ids, id)
        }

        _ = json.NewEncoder(w).Encode(ids)
    }).Methods("GET")

    router.HandleFunc("/p2p/ratings/publish/{type}", func(w http.ResponseWriter, r *http.Request) {
        contract := new(models.Contract)
        publishType := mux.Vars(r)["type"]

        var rating *Rating

        if publishType == "vendor" {
            rating = VendorRatingFromContract(contract)
        } else if publishType == "buyer" {
            rating = BuyerRatingFromContract(contract)
        } else {
            _ = json.NewEncoder(w).Encode(map[string]interface{}{
                "error": "endpoint only accepts either 'vendor' or 'buyer'",
            })
            return
        }

        var errCode string
        b, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Debugf("failed to read body: %v", err)
        }

        _ = json.Unmarshal(b, &contract)

        if err != nil {
            log.Debugf("failed to marshal json: %v\n%v", err, string(b))
            errCode = fmt.Sprintf("failed to marshal json: %v", err)
        } else {
            err := skademlia.Broadcast(sat.Node, satellite.Packet{
                PacketType: satellite.PType_Broadcast,
                Namespace:  "new_rating",
                Payload:    rating,
            })
            if err != nil {
                log.Debugf("failed to broadcast: %v", err)
                errCode = fmt.Sprintf("failed to broadcast: %v", err)
            }
        }

        _ = json.NewEncoder(w).Encode(map[string]interface{}{
            "error": errCode,
        })
    })

    router.HandleFunc("/p2p/ratings/get/{peer}/{ids}", func(w http.ResponseWriter, r *http.Request) {

        vars := mux.Vars(r)
        var errCode string
        var ratings []interface{}

        p, exists := sat.Peers[vars["peer"]]
        if exists {
            start := time.Now()
            rs, err := sat.Request(p, "get_rating", RatingRequest{vars["ids"]})
            if err != nil {
                log.Errorf("failed to write: %v", err)
                errCode = fmt.Sprintf("failed to write: %v", err)
            } else {
                log.Debug("Waiting for streams")
                for inbound := range rs.Stream {
                    ratings = append(ratings, inbound.Payload)
                }
            }
            log.Debug("Waiting for streams is complete: ", time.Now().Sub(start))
        } else {
            errCode = fmt.Sprintf("peer does not exist: %v", vars["peer"])
        }

        _ = json.NewEncoder(w).Encode(map[string]interface{}{
            "ratings": ratings,
            "error":   errCode,
        })
    })

    router.HandleFunc("/p2p/ratings/seek-sync/{ids}", func(w http.ResponseWriter, r *http.Request) {

        vars := mux.Vars(r)
        var errCode string
        var ratings []interface{}

        rs, err := sat.Seek("get_rating", RatingRequest{vars["ids"]})
        if err != nil {
            log.Errorf("failed to broadcast: %v", err)
            errCode = fmt.Sprintf("failed to write: %v", err)
        } else {
            log.Debug("Waiting for streams")
            for inbound := range rs.Stream {
                ratings = append(ratings, inbound.Payload)
            }
        }

        _ = json.NewEncoder(w).Encode(map[string]interface{}{
            "ratings": ratings,
            "error":   errCode,
        })
    })

    return router
}
