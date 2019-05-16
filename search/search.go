package search

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicelogger"

	"github.com/robertkrimen/otto"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
)

type JSManager struct {
	Vm  *otto.Otto
	Log *servicelogger.LogPrinter
}

func (sm *JSManager) Compare(document string, qstub string) (bool, error) {
	// TODO do some JSON reencoding first to prevent Query Injection vulnerability.
	// err := sm.Vm.Set("document", document)
	// if err != nil {
	// 	sm.Log.Error(fmt.Sprintf("Failed to import document: %v", err))
	// }
	code := fmt.Sprintf(`%v.test(%v)`, qstub, document)
	vmval, err := sm.Vm.Run(code)
	if err != nil {
		return false, fmt.Errorf("Engine Error: %v \n--value--\n%v\n--code--\n%v", err, vmval, code)
	}
	// val, err := sm.Vm.Get("result")
	// if err != nil {
	// 	return false, fmt.Errorf("Query Error: %v", err)
	// }
	// sm.Log.Debug(fmt.Sprintf("\n--value--\n%v\n--code--\n%v\n", vmval, code))
	result, _ := vmval.ToBoolean()
	return result, nil
}

func (sm *JSManager) CreateQueryStub(signature string, query string) error {
	code := fmt.Sprintf(`%v = new mingo.Query(%v)`, signature, query)
	_, err := sm.Vm.Run(code)
	// str, _ := res.ToString()
	return err
}

func loadToVM(log *servicelogger.LogPrinter, vm *otto.Otto, filename string) error {
	log.Debug("Loading Javascript: " + filename)
	jscode, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to load: %v", filename))
		return fmt.Errorf("Failed to load: %v", filename)
	}

	vm.Run(`mingo = 'Mingo is uninitialized'`)
	val, err := vm.Run(string(jscode))
	if err != nil {
		log.Error(fmt.Sprintf("VM Err: %v", err))
	} else {
		log.Debug(fmt.Sprintf("VM File: %v", val))
	}
	return nil
}

func InitializeJSVM(log *servicelogger.LogPrinter) (*JSManager, error) {
	log.Debug("Loading Search Manager")
	sm := JSManager{
		Vm:  otto.New(),
		Log: log,
	}
	for _, file := range []string{"external/babel.js", "external/polyfills.js", "external/mingo.js"} {
		loadToVM(log, sm.Vm, file)
	}
	// sm.Vm.Run(`__querystubs = {}`)
	return &sm, nil
}

type QueryPacket struct {
	Packet    *models.Listing
	QueryStub string
	Result    chan *models.Listing
}

type QueryEngine struct {
	Managers       []*JSManager
	Log            *servicelogger.LogPrinter
	PendingPackets chan *QueryPacket
}

func QMuxWorker(id int, vm *JSManager, packetStream chan *QueryPacket) {
	for packet := range packetStream {
		//fmt.Printf("[Worker-%v] Processing %v\n", id, packet.Packet.Title)
		data, _ := json.Marshal(packet.Packet)
		result, _ := vm.Compare(string(data), packet.QueryStub)
		if result {
			packet.Result <- packet.Packet
		} else {
			packet.Result <- nil
		}
	}
}

func (qe *QueryEngine) QueryMultiplexer() {
	qe.Log.Info("Running Multiplexer")
	for idx, sm := range qe.Managers {
		go QMuxWorker(idx, sm, qe.PendingPackets)
	}
}

func (qe *QueryEngine) QueryListings(collection []*models.Listing, qstub string) []*models.Listing {
	results := []*models.Listing{}
	COLLECTION_COUNT := len(collection)
	res := make(chan *models.Listing, COLLECTION_COUNT)

	for _, listing := range collection {
		qe.Log.Debug(fmt.Sprintf("Sending: %v\n", listing.Title))
		qe.PendingPackets <- &QueryPacket{Packet: listing, QueryStub: qstub, Result: res}
	}

	qe.Log.Debug(fmt.Sprintf("Waiting for the results..."))
	for a := 1; a <= COLLECTION_COUNT; a++ {
		result := <-res
		qe.Log.Debug(fmt.Sprintf("[%v]RESULT: %v", a, result))
		if result != nil {
			results = append(results, result)
		}
	}

	return results
}

func (qe *QueryEngine) CreateQueryStub(signature string, query string) error {
	for _, sm := range qe.Managers {
		code := fmt.Sprintf(`%v = new mingo.Query(%v)`, signature, query)
		_, err := sm.Vm.Run(code)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitializeQueryEngine(log *servicelogger.LogPrinter, workers int) *QueryEngine {
	log.Info("Initializing Query Engine")
	qe := QueryEngine{}
	qe.Log = log
	qe.PendingPackets = make(chan *QueryPacket, 1000)
	count := make(chan bool, workers)
	for i := 0; i <= workers; i++ {
		go func() {
			se, _ := InitializeJSVM(log)
			qe.Managers = append(qe.Managers, se)
			count <- true
		}()
	}
	for i := 0; i <= workers; i++ {
		<-count
	}
	qe.QueryMultiplexer()
	return &qe
}

// Find the listings and returns potential matches via supplied keyword
func Find(keyword string, averageRating int64, listings []*models.Listing) []*models.Listing {
	fmt.Println(keyword)
	response := []*models.Listing{}
	for _, listing := range listings {
		if findByKeyword(keyword, listing) && findByAverageRating(averageRating, listing) {
			response = append(response, listing)
		}
	}
	return response
}

func findByKeyword(keyword string, listing *models.Listing) bool {
	keywordLowercase := strings.ToLower(keyword)
	// Probably an initial wildcard search or just browsing via filters
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(listing.Title), keywordLowercase) || strings.Contains(strings.ToLower(listing.Description), keywordLowercase)
}

func findByAverageRating(averageRating int64, listing *models.Listing) bool {
	if averageRating <= 0 {
		return true
	}
	return listing.AverageRating >= averageRating
}
