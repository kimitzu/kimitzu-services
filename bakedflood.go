package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/location"
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

	go voyager.RunVoyagerService(srvLog.Spawn("voyager"), store)
	location.RunLocationService(srvLog.Spawn("location"))

}
