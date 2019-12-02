package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/go-homedir"
	"github.com/nokusukun/particles/config"
	"github.com/nokusukun/particles/roggy"

	"github.com/kimitzu/kimitzu-services/api"
	"github.com/kimitzu/kimitzu-services/configs"
	"github.com/kimitzu/kimitzu-services/p2p"

	"github.com/kimitzu/kimitzu-services/location"
	"github.com/kimitzu/kimitzu-services/servicestore"
	"github.com/kimitzu/kimitzu-services/voyager"
)

var (
	logger     = roggy.Printer("services")
	confSat    = config.Satellite{}
	confDaemon = configs.Daemon{}
)

func init() {
	flag.UintVar(&confSat.Port, "p2p-port", 9009, "Listen for peers in specified port")
	flag.StringVar(&confSat.Host, "p2p-host", "0.0.0.0", "Listen for peers in this host")
	flag.BoolVar(&confSat.DisableUPNP, "p2p-noupnp", false, "disable UPNP")

	//flag.StringVar(&confDaemon.ArgLogLevel, "log", "0", "Port to listen")
	flag.StringVar(&confDaemon.DataPath, "data", "&home", "Folder to store location data")

	flag.StringVar(&confDaemon.DialTo, "dial", "", "Bootstrap s/kad from this peer")
	flag.StringVar(&confDaemon.BootstrapNodeIdentity, "bootstrapNodeIdentity", "", "The identity (host:port) of this bootstrap node")
	flag.StringVar(&confDaemon.ApiListen, "api", "0.0.0.0:8109", "Enable the api and serve to this address")
	flag.StringVar(&confDaemon.DatabasePath, "dbpath", "&home", "Database Path")
	flag.StringVar(&confDaemon.KeyPath, "key", "&home", "Read/write key from/to path")
	flag.BoolVar(&confDaemon.GenerateNewKeys, "generate", true, "Generate new keys")
	flag.BoolVar(&confDaemon.ShowHelp, "h", false, "Show help")
	flag.IntVar(&roggy.LogLevel, "log", 2, "log level 0~5")

	flag.Parse()

	if confDaemon.DataPath == "&home" {
		home, _ := homedir.Dir()
		confDaemon.DataPath = path.Join(home, "djali")
	}

	if confDaemon.DatabasePath == "&home" {
		confDaemon.DatabasePath = path.Join(confDaemon.DataPath, "p2p")
	}

	if confDaemon.KeyPath == "&home" {
		confDaemon.KeyPath = path.Join(confDaemon.DataPath, "p2pkeys")
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
	log.Infof("Log Level: %v", roggy.LogLevel)
	log.Info("Starting Services")

	store := servicestore.InitializeManagedStorage(confDaemon.DataPath)
	p2pKillSig := make(chan int, 1)

	// database initialization
	ratingManager, err := p2p.InitializeRatingManager(confDaemon.DatabasePath)
	if err != nil {
		log.Error("Opening database failed")
		roggy.Wait()
		panic(err)
	}

	// test(&srvLog, log, store)
	apiRouter := mux.NewRouter()

	time.Sleep(time.Second * 10)
	go p2p.Bootstrap(&confDaemon, &confSat, ratingManager, p2pKillSig)
	go voyager.RunVoyagerService(log.Sub("voyager"), store)
	location.RunLocationService(log.Sub("location"))

	p2p.AttachAPI(p2p.Sat, apiRouter, ratingManager)
	api.AttachStore(store)
	api.AttachAPI(log.Sub("api"), apiRouter)

	log.Infof("Running API on %v", confDaemon.ApiListen)
	log.Error(http.ListenAndServe(confDaemon.ApiListen, apiRouter))

	select {}
}
