package main

import (
	"fmt"

	"gitlab.com/kingsland-team-ph/djali/djali-services/location"
)

func main() {
	fmt.Println("BakedLasagna Loader")
	location.RunLocationService()
}
