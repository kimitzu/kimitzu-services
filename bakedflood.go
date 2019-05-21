package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/location"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/search"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicelogger"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicestore"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/voyager"
)

var (
	ArgLogLevel string
	LogLevel    int
	Version     string
)

func init() {
	flag.StringVar(&ArgLogLevel, "log", "0", "Port to listen")
	flag.Parse()
	llvl, _ := strconv.Atoi(ArgLogLevel)
	LogLevel = llvl
	Version = "0.0.1"
}

func lookupTest(log *servicelogger.LogPrinter, filter string, searchEngine *search.QueryEngine, store *servicestore.MainStorage) int {
	param := &search.QueryParameters{
		Collection: store.Listings,
		Limit:      10,
		Query:      filter,
	}
	result := searchEngine.QueryListings(param)
	log.Info(fmt.Sprintf("Returned: %v", result))
	for idx, listing := range result.Result {
		log.Verbose(fmt.Sprintf("[%v]Name: %v", idx, listing.Title))
	}
	return len(result.Result)
}

func test(srvLog *servicelogger.LogManager, log *servicelogger.LogPrinter, store *servicestore.MainStorage) {
	queryEngine := search.InitializeQueryEngine(srvLog.Spawn("explorer"), 20)
	// voyager.Initialize(srvLog.Spawn("voyager-init"), store)
	log.Info(fmt.Sprintf("Store has %v entires.", len(store.Listings)))
	log.Info("Starting Lookup...")

	timeflags := make(chan bool, 2)

	go func() {
		mstart := time.Now()
		count := lookupTest(log, `function(doc) {
			return doc.price.currencyCode === 'USD'
		}`, queryEngine, store)
		msend := time.Now()
		timeflags <- true
		fmt.Printf(fmt.Sprintf("Price Code Results: found %v items in %v\n", count, msend.Sub(mstart)))
	}()

	go func() {
		time.Sleep(time.Second * 1)
		amstart := time.Now()
		acount := lookupTest(log, `function(doc) {
			return doc.price.currencyCode === 'USD' && doc.price.amount > 1000
		}`, queryEngine, store)
		amsend := time.Now()
		timeflags <- true
		fmt.Printf(fmt.Sprintf("USD Code Results: found %v items in %v\n", acount, amsend.Sub(amstart)))
	}()

	start := time.Now()
	<-timeflags
	<-timeflags
	end := time.Now()
	fmt.Printf(fmt.Sprintf("Total elapsed time: %v\n", end.Sub(start)))
}

func main() {
	srvLog := servicelogger.LogManager{}
	go srvLog.Start(LogLevel)

	// Deadlock prevention
	time.Sleep(time.Second * 1)

	log := srvLog.Spawn("ServiceLoader")
	log.Info(fmt.Sprintf("Djali Services Daemon (%v)", Version))
	log.Info(" --- --- --- --- --- ")
	log.Info("Log Level: " + ArgLogLevel)
	log.Info("Starting Services")

	store := servicestore.InitializeManagedStorage()

	// test(&srvLog, log, store)
	time.Sleep(time.Second * 10)

	go voyager.RunVoyagerService(srvLog.Spawn("voyager"), store)
	location.RunLocationService(srvLog.Spawn("location"))

}
