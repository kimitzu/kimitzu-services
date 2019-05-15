package main

import (
	"fmt"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/location"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/servicelogger"
	"gitlab.com/kingsland-team-ph/djali/djali-services.git/voyager"
)

func main() {
	fmt.Println("BakedLasagna Loader")
	srvLog := servicelogger.LogManager{}
	go srvLog.Start(5)

	fmt.Println("Running Services")
	go voyager.RunVoyagerService(srvLog.Spawn("voyager"))
	location.RunLocationService(srvLog.Spawn("location"))
}
