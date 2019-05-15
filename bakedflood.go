package main

import (
	"flag"
	"fmt"
	"strconv"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/location"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicelogger"
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

	log := srvLog.Spawn("ServiceLoader")
	log.Info(fmt.Sprintf("Djali Services Daemon (%v)", Version))
	log.Info(" --- --- --- --- --- ")
	log.Info("Log Level: " + ArgLogLevel)
	log.Info("Starting Services")

	go voyager.RunVoyagerService(srvLog.Spawn("voyager"))
	location.RunLocationService(srvLog.Spawn("location"))
}
