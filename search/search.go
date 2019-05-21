package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicelogger"

	"github.com/robertkrimen/otto"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
)

type JSManager struct {
	ID        string
	Vm        *otto.Otto
	Log       *servicelogger.LogPrinter
	Interrupt func()
	Nuked     bool
}

// Clone - Clones the VM context of the manager, returns a new copy.
func (sm *JSManager) CloneMod(command string) *JSManager {
	jsm := &JSManager{
		Vm:  sm.Vm.Copy(),
		Log: sm.Log,
	}
	jsm.Vm.Run(command)
	jsm.AttachInterrupt()
	return jsm
}

// AttachInterrupt - Attaches an interrupt channel to kill the VM
func (sm *JSManager) AttachInterrupt() {
	sm.Vm.Interrupt = make(chan func(), 1)
	sm.Nuked = false
}

// Nuke - Nukes the VM by sending a panic to the VM
func (sm *JSManager) Nuke() {
	sm.Vm.Interrupt <- func() {
		panic(errors.New("Kill Request"))
	}
	sm.Nuked = true
}

// Compare - Compares a json string document with a qstub
func (sm *JSManager) Compare(document string) (bool, error) {
	code := fmt.Sprintf(`q(%v)`, document)
	//code := ``
	vmval, err := sm.Vm.Run(code)
	if err != nil {
		return false, fmt.Errorf("Engine Error: %v \n--value--\n%v\n--code--\n%v", err, vmval, code)
	}
	result, _ := vmval.ToBoolean()
	return result, nil
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

// InitializeJSVM - Spins up a new VM instance
func InitializeJSVM(id string, log *servicelogger.LogPrinter) (*JSManager, error) {
	log.Debug("Loading Search Manager")
	sm := JSManager{
		ID:  id,
		Vm:  otto.New(),
		Log: log,
	}
	for _, file := range []string{"external/babel.js", "external/polyfills.js", "external/mingo.js"} {
		loadToVM(log, sm.Vm, file)
	}
	sm.AttachInterrupt()
	return &sm, nil
}

type QueryEngine struct {
	ParentVM  *JSManager
	ParkedVMs chan *JSManager
	Log       *servicelogger.LogPrinter
}

type QueryParameters struct {
	Collection  []*models.Listing `json:"-"`
	Query       string            `json:"query"`
	Limit       int               `json:"limit"`
	Start       int               `json:"start"`
	WorkerCount int               `json:"-"`
}

type QueryResponse struct {
	Result []*models.Listing `json:"result"`
	End    int               `json:"end"`
}

func QMuxWorker(id int, vm *JSManager, collections []*models.Listing, results chan *models.Listing) {
	for _, packet := range collections {
		//fmt.Printf("[Worker-%v] Processing %v\n", id, packet.Packet.Title)
		data, _ := json.Marshal(packet)
		result, _ := vm.Compare(string(data))
		if result {
			results <- packet
		} else {
			results <- nil
		}
	}
}

//func (qe *QueryEngine) QueryListings(collection []*models.Listing, query string, limit int) []*models.Listing {
func (qe *QueryEngine) QueryListings(parameters *QueryParameters) *QueryResponse {
	workerCount := 2
	if parameters.WorkerCount != 0 {
		workerCount = parameters.WorkerCount
	}
	query := parameters.Query

	queryStart := 0
	if parameters.Start != 0 {
		queryStart = parameters.Start
	}

	if queryStart > len(parameters.Collection) {
		return nil
	}
	queryEnd := queryStart
	collection := parameters.Collection[queryStart:len(parameters.Collection)]
	COLLECTION_COUNT := len(collection)
	
	queryLimit := len(collection)
	if parameters.Limit != 0 {
		if queryLimit > COLLECTION_COUNT {
			queryLimit = COLLECTION_COUNT
		} else {
			queryLimit = parameters.Limit
		}
	}

	results := []*models.Listing{}

	res := make(chan *models.Listing, COLLECTION_COUNT)
	pendingPackets := make(chan *models.Listing, 1000)
	workers := []*JSManager{}

	s1 := time.Now()
	qe.Log.Verbose("Acquiring Parked VMs")
	// Create The VMs
	for i := 1; i <= workerCount; i++ {
		vm := <-qe.ParkedVMs
		qe.Log.Verbose(fmt.Sprintf("Acquired parked vm: %v", vm.ID))
		vm.Vm.Run(fmt.Sprintf(`q = %v`, query))
		workers = append(workers, vm)
	}
	e1 := time.Now()
	qe.Log.Verbose(fmt.Sprintf("AcquireVM: %v", e1.Sub(s1)))

	// Defer to make sure that the VM gets returned to the VMBuffer
	defer func() {
		s5 := time.Now()
		qe.Log.Verbose("Parking Used VMs...")
		go func() {
			close(pendingPackets)
			for _, vm := range workers {
				qe.ParkedVMs <- vm
			}
			e5 := time.Now()
			qe.Log.Verbose(fmt.Sprintf("Park: %v", e5.Sub(s5)))
		}()
	}()

	s2 := time.Now()

	qe.Log.Verbose("Spinning up Multiplexer...")
	// Spin up the QueryMultiplexer
	chunkSize := COLLECTION_COUNT / workerCount
	for idx, sm := range workers {
		start := idx * chunkSize
		end := (idx + 1) * chunkSize
		remainder := COLLECTION_COUNT - end
		if remainder < end-start {
			qe.Log.Verbose("Sending the last bits")
			end += COLLECTION_COUNT - end
		}
		chunk := collection[start:end]
		qe.Log.Verbose(fmt.Sprintf("Chunk Range: [%v:%v] Rem: %v", start, end, remainder))
		go QMuxWorker(idx, sm, chunk, res)
	}
	e2 := time.Now()
	qe.Log.Verbose(fmt.Sprintf("SpinupMultiplex: %v", e2.Sub(s2)))

	s3 := time.Now()
	qe.Log.Verbose("Sending queries...")
	// Send the Collection
	// for _, listing := range collection {
	// 	pendingPackets <- &QueryPacket{Packet: listing, Result: res}
	// }
	e3 := time.Now()
	qe.Log.Verbose(fmt.Sprintf("SendQuery: %v", e3.Sub(s3)))

	s4 := time.Now()
	qe.Log.Verbose("Gathering Results...")
	// Retrieve results
	qe.Log.Debug(fmt.Sprintf("Waiting for the results..."))
	for a := 1; a <= COLLECTION_COUNT; a++ {
		result := <-res
		qe.Log.Debug(fmt.Sprintf("Recieving [%v]", a))
		if result != nil {
			results = append(results, result)
			if len(results) >= queryLimit {
				break
			}
		}
		queryEnd++
	}
	e4 := time.Now()
	qe.Log.Verbose(fmt.Sprintf("GetResult: %v", e4.Sub(s4)))

	// return results, queryEnd
	return &QueryResponse{
		Result: results,
		End:    queryEnd,
	}
}

func InitializeQueryEngine(log *servicelogger.LogPrinter, workers int) *QueryEngine {
	const VMCOUNT = 6
	log.Info("Initializing Query Engine")
	qe := QueryEngine{}
	qe.ParkedVMs = make(chan *JSManager, VMCOUNT)
	qe.Log = log

	// Create Managers
	for i := 1; i <= VMCOUNT; i++ {
		vm, _ := InitializeJSVM(fmt.Sprintf("jsvm-%v", i), log)
		qe.ParkedVMs <- vm
	}

	// qe.PendingPackets = make(chan *QueryPacket, 1000)
	// se, _ := InitializeJSVM(log)
	// qe.ParentVM = se

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
