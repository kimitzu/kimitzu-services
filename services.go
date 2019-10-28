package main

import (
	"flag"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/nokusukun/particles/config"
	"github.com/nokusukun/particles/roggy"

	"github.com/djali-foundation/djali-services/api"
	"github.com/djali-foundation/djali-services/configs"
	"github.com/djali-foundation/djali-services/p2p"

	"github.com/djali-foundation/djali-services/location"
	"github.com/djali-foundation/djali-services/servicestore"
	"github.com/djali-foundation/djali-services/voyager"
)

var (
	logger     = roggy.Printer("particled")
	confSat    = config.Satellite{}
	confDaemon = configs.Daemon{}
)



func init() {
	flag.UintVar(&confSat.Port, "p2p-port", 3000, "Listen for peers in specified port")
	flag.StringVar(&confSat.Host, "p2p-host", "127.0.0.1", "Listen for peers in this host")
	flag.BoolVar(&confSat.DisableUPNP, "p2p-noupnp", false, "disable UPNP")

	flag.StringVar(&confDaemon.ArgLogLevel, "log", "0", "Port to listen")
	flag.StringVar(&confDaemon.DataPath, "data", "&home", "Folder to store location data")

	flag.StringVar(&confDaemon.DialTo, "dial", "", "Bootstrap s/kad from this peer")
	flag.StringVar(&confDaemon.ApiListen, "api", "", "Enable the api and serve to this address")
	flag.StringVar(&confDaemon.DatabasePath, "dbpath", "", "Database Path")
	flag.StringVar(&confDaemon.KeyPath, "key", "", "Read/write key from/to path")
	flag.BoolVar(&confDaemon.GenerateNewKeys, "generate", false, "Generate new keys")
	flag.BoolVar(&confDaemon.ShowHelp, "h", false, "Show help")
	flag.IntVar(&roggy.LogLevel, "log", 2, "log level 0~5")

	flag.Parse()

	if confDaemon.DataPath == "&home" {
		home, _ := homedir.Dir()
		confDaemon.DataPath = path.Join(home, "djali")
	}

	confDaemon.Version = "0.1.2"

}

func main() {
	log.SetFlags(0) // Disables internal logging
	log := logger

	// Deadlock prevention
	time.Sleep(time.Second * 1)

	log.Info(fmt.Sprintf("Djali Services Daemon (%v)", confDaemon.Version))
	log.Info(" --- --- --- --- --- ")
	log.Info("Log Level: " + confDaemon.ArgLogLevel)
	log.Info("Starting Services")

	store := servicestore.InitializeManagedStorage(confDaemon.DataPath)

	// test(&srvLog, log, store)
	time.Sleep(time.Second * 10)
	go p2p.Bootstrap(&confDaemon, &confSat)
	go voyager.RunVoyagerService(log.Sub("voyager"), store)
	location.RunLocationService(log.Sub("location"))
	api.AttachStore(store)
	api.RunHTTPService(log.Sub("api"))
}
