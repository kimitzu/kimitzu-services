package p2p

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/nokusukun/particles/config"
	"github.com/nokusukun/particles/keys"
	"github.com/nokusukun/particles/roggy"
	"github.com/nokusukun/particles/satellite"
	"github.com/perlin-network/noise/skademlia"

	"github.com/kimitzu/kimitzu-services/configs"
)

var log = roggy.Printer("p2p")
var Sat *satellite.Satellite

func Bootstrap(cdae *configs.Daemon, csat *config.Satellite, ratingManager *RatingManager, killsig chan int) {
	var boostrapNodes = [1]string{"109.201.140.20:9009"}

	//todo: move most of this in the services main function
	log.Info("Starting Particle Daemon")
	printSplash()

	log.Debug(roggy.Clr("TURNING ON DEBUG LOGS WILL SEVERELY IMPACT PERFORMANCE", 1))

	if cdae.DatabasePath == "" {
		log.Error("No database path provided --dbpath")
		roggy.Wait()
		os.Exit(1)
	}

	// satellite bootstrapping
	keyPair, err := getKeys(cdae.KeyPath, cdae.GenerateNewKeys)
	if err != nil {
		log.Error("Failed to get keyPair:", err)
		log.Error("Your key might not exist, try with the -generate flag")
	}
	Sat = satellite.BuildNetwork(csat, keyPair)

	for i := 0; i < len(boostrapNodes); i++ {
		if cdae.BootstrapNodeIdentity != boostrapNodes[i] {
			dial(boostrapNodes[i])
		}
	}

	if cdae.DialTo != "" {
		dial(cdae.DialTo)
	}

	bootstrapEvents(Sat, ratingManager)

	// API
	//if cdae.ApiListen != "" {
	//    log.Notice("Starting API on:", cdae.ApiListen)
	//    router := generateAPI(sat)
	//    log.Error(http.ListenAndServe(cdae.ApiListen, router))
	//} else {
	//    log.Notice("No API port provided")
	//}

	defer func() {
		err := ratingManager.Close()
		if err != nil {
			log.Error("failed to close database", err)
		}
		log.Info("Killing node...")
		Sat.Node.Kill()
		roggy.Wait()
	}()

	<-killsig
}

func dial(node string) {
	log.Info("Connecting s/kad bootstrap at ", node)
	peer, err := Sat.Node.Dial(node)
	if err != nil {
		log.Errorf("Failed to dial to s/kad bootstrap")
	}
	log.Debugf("waiting %v for bootstrap s/kad authentication", node)
	skademlia.WaitUntilAuthenticated(peer)
	log.Infof("Bootstrapped to: %v", satellite.GetPeerID(peer))
}

func getKeys(path string, newKeys bool) (*skademlia.Keypair, error) {
	_, err := os.Stat(path)
	if err == nil {
		log.Notice("-generate flag specified but key already exists, using that instead")
	}

	if newKeys && err != nil {

		log.Info("Generating new keys...")
		newkeys := skademlia.RandomKeys()
		kb, err := keys.Serialize(newkeys)
		if err != nil {
			panic(err)
		}

		if path == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter key filename: ")
			path, err = reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
		}

		log.Infof("New key generated: %v", hex.EncodeToString(newkeys.PublicKey()))
		err = ioutil.WriteFile(path, kb, os.ModePerm)
		if err != nil {
			panic(err)
		}

		log.Infof("Key saved to: %v", path)
	}

	return keys.ReadKeys(path)
}

func printSplash() {
	fmt.Print(roggy.Clr(`
                                      I8                    ,dPYb,                  8I 
                                      I8                    IP''Yb                  8I 
                                   88888888  gg             I8  8I                  8I 
                                      I8     ""             I8  8'                  8I 
 gg,gggg,      ,gggg,gg   ,gggggg,    I8     gg     ,gggg,  I8 dP   ,ggg,     ,gggg,8I 
 I8P"  "Yb    dP"  "Y8I   dP""""8I    I8     88    dP"  "Yb I8dP   i8" "8i   dP"  "Y8I 
 I8'    ,8i  i8'    ,8I  ,8'    8I   ,I8,    88   i8'       I8P    I8, ,8I  i8'    ,8I 
,I8 _  ,d8' ,d8,   ,d8b,,dP     Y8, ,d88b, _,88,_,d8,_    _,d8b,_  'YbadP' ,d8,   ,d8b,
PI8 YY88888PP"Y8888P"'Y88P      'Y888P""Y888P""Y8P""Y8888PP8P'"Y88888P"Y888P"Y8888P"'Y8
 I8                                                                                    
 ?`, roggy.LogLevel))
	fmt.Print(roggy.Clr(fmt.Sprintf("\t[ Particle Daemon running on log level %v ]\n", roggy.LogLevel), roggy.LogLevel))
}
