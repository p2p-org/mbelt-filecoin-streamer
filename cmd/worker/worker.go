package main

import (
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"log"
)

const (
	defaultHeight = 5000
)

var conf *config.Config

func init() {
	/*
		conf = &config.Config{
			APIUrl:     "ws://116.203.240.62:1234/rpc/v0",
			APIToken:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.3b1-x0KwOB1-60NzHNGsMdMyzr6b0kokzh-Z4bc400Y",
			KafkaHosts: "localhost:9092",
		}*/

	conf = &config.Config{
		APIUrl:     "ws://127.0.0.1:1234/rpc/v0",
		KafkaHosts: "localhost:9092",
	}
}

func main() {
	var syncHeight abi.ChainEpoch
	err := services.InitServices(conf)

	if err != nil {
		log.Println("[App][Debug]", "Cannot init services:", err)
		return
	}

	head := services.App().BlocksService().GetHead()

	if head != nil && head.Height() > 0 {
		syncHeight = head.Height()
	} else {
		log.Println("[App][Debug]", "Cannot get header, use default syncHeight:", defaultHeight)
		syncHeight = defaultHeight
	}

	for height := abi.ChainEpoch(0); height < syncHeight; height++ {
		log.Println("[Datastore][Debug]", "Load height:", height)

		tipSet, isCanContinue := services.App().BlocksService().GetByHeight(height)

		if !isCanContinue {
			log.Println("[App][Debug]", "Height reached")
			return
		}

		// Empty TipSet, skipping
		if tipSet == nil {
			continue
		}

		services.App().BlocksService().Push(tipSet.Blocks())

		for _, block := range tipSet.Blocks() {
			if block.Messages.Defined() {
				messages := services.App().MessagesService().GetBlockMessages(block.Messages)

				if messages == nil {
					continue
				}

				if len(messages.Cids) > 0 {
					for _, messageCid := range messages.Cids {
						message := services.App().MessagesService().GetMessage(messageCid)

						if message == nil {
							continue
						}

						services.App().MessagesService().Push(message)
					}
				}
			}
		}

	}

}
