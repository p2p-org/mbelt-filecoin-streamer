package main

import (
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"log"
)

const (
	// Testnet node
	urlDev     = "ws://116.203.240.62:1234/rpc/v0"
	originLDev = "http://localhost/"
	jwtDev     = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.3b1-x0KwOB1-60NzHNGsMdMyzr6b0kokzh-Z4bc400Y"

	// Localhost (next) node
	urlLocal    = "ws://127.0.0.1:1234/rpc/v0"
	originLocal = "http://localhost/"
	jwtLocal    = ""
)

var (
	apiClient *client.APIClient
	ds        *datastore.Datastore
)

func main() {
	var err error

	//apiClient, err = client.Init(urlDev, originLDev, jwtDev)
	apiClient, err = client.Init(urlLocal, originLocal, jwtLocal)

	if err != nil {
		return
	}

	ds, err = datastore.Init("localhost:9092")

	if err != nil {
		return
	}

	getData()

}

func getData() {
	head := apiClient.GetHead()

	if head == nil {
		log.Println("[Error]", "Cannot get head")
		return
	}

	maxHeight := head.Height()

	if maxHeight == 0 {
		log.Println("[Error]", "Empty chain")
		return
	}

	for height := abi.ChainEpoch(0); height < maxHeight; height++ {
		tipSet := apiClient.GetByHeight(height)

		// log.Println("Height", height)

		if tipSet == nil {
			// log.Println("[Error]", "Empty tipSet")
			continue
		}

		// log.Println("Blocks count", len(tipSet.Blocks()))
		for _, block := range tipSet.Blocks() {

			// <- chan store push Block data

			/*
				blockMessages := apiClient.GetBlockMessages(block.Messages)
				if blockMessages == nil  {
					// log.Println("[Error]", "Empty blockMessages")
					continue
				}

				// Process BlsMessages, SecpkMessages if returns
			*/

			message := apiClient.GetMessage(block.Messages)
			if message != nil {
				// <- chan store push message data

				if !message.From.Empty() {
					log.Printf("%x\n", message)
				}
			}

		}
	}

	log.Println("Finish")
}
