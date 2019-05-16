package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

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

func lookupTest(log *servicelogger.LogPrinter, stubname string, searchEngine *search.QueryEngine, store *servicestore.MainStorage) int {
	result := searchEngine.QueryListings(store.Listings, stubname)
	return len(result)
}

func test(srvLog *servicelogger.LogManager, log *servicelogger.LogPrinter, store *servicestore.MainStorage) {
	queryEngine := search.InitializeQueryEngine(srvLog.Spawn("explorer"), 20)
	voyager.Initialize(srvLog.Spawn("voyager-init"), store)
	log.Info(fmt.Sprintf("Store has %v entires.", len(store.Listings)))
	log.Info("Starting Lookup...")
	queryEngine.CreateQueryStub("asdsdaswe", `{$and: [{"price.amount": {$gt: 1000}}, {"price.currencyCode": "USD"}]}`)
	queryEngine.CreateQueryStub("asdsaasda", `{"price.currencyCode": "USD"}`)

	go func() {
		mstart := time.Now()
		count := lookupTest(log, "asdsdaswe", queryEngine, store)
		msend := time.Now()
		fmt.Printf(fmt.Sprintf("Results: found %v items in %v\n", count, msend.Sub(mstart)))
	}()

	time.Sleep(time.Second * 1)
	amstart := time.Now()
	acount := lookupTest(log, "asdsaasda", queryEngine, store)
	amsend := time.Now()

	fmt.Printf(fmt.Sprintf("Results: found %v items in %v\n", acount, amsend.Sub(amstart)))
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

	store := servicestore.InitializeStore()

	test(&srvLog, log, store)

	// go voyager.RunVoyagerService(srvLog.Spawn("voyager"), store)
	// location.RunLocationService(srvLog.Spawn("location"))

}
