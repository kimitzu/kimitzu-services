package main

import (
	"fmt"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/location"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/voyager"
)

type ServicesLog struct {
	LocationService chan string
	VoyagerService  chan string
}

func (l *ServicesLog) Init() {
	l.LocationService = make(chan string, 100)
	l.VoyagerService = make(chan string, 100)
}

func (l *ServicesLog) Run() {
	for {
		select {
		case loc := <-l.LocationService:
			fmt.Println("[Location] " + loc)
		case loc := <-l.VoyagerService:
			fmt.Println("[Voyager] " + loc)
		}
	}
}

func main() {
	fmt.Println("BakedLasagna Loader")
	srvLog := ServicesLog{}
	srvLog.Init()
	go srvLog.Run()

	go voyager.RunVoyagerService(srvLog.VoyagerService)
	location.RunLocationService(srvLog.LocationService)
}
