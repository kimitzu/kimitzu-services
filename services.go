package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/djali-foundation/djali-services/api"

	"github.com/djali-foundation/djali-services/location"
	"github.com/djali-foundation/djali-services/servicelogger"
	"github.com/djali-foundation/djali-services/servicestore"
	"github.com/djali-foundation/djali-services/voyager"
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

func main() {
	log.SetFlags(0) // Disables internal logging
	srvLog := servicelogger.LogManager{}
	go srvLog.Start(LogLevel)

	// Deadlock prevention
	time.Sleep(time.Second * 1)

	log := srvLog.Spawn("ServiceLoader")
	log.Info(fmt.Sprintf("Djali Services Daemon (%v)", Version))
	log.Info(" --- --- --- --- --- ")
	log.Info("Log Level: " + ArgLogLevel)
	log.Info("Starting Services")

	store := servicestore.InitializeManagedStorage("")

	// test(&srvLog, log, store)
	time.Sleep(time.Second * 10)

	go voyager.RunVoyagerService(srvLog.Spawn("voyager"), store)
	location.RunLocationService(srvLog.Spawn("location"))
	api.AttachStore(store)
	api.RunHTTPService(srvLog.Spawn("api"))

}
