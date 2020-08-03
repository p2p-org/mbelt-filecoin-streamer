package main

import (
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"log"
	"os"
	"strconv"
)

const (
	defaultHeight = 5000
)

var conf *config.Config

func init() {
	conf = &config.Config{
		APIUrl:     os.Getenv("MBELT_FILECOIN_STREAMER_API_URL"),
		APIToken:   os.Getenv("MBELT_FILECOIN_STREAMER_API_TOKEN"),
		KafkaHosts: os.Getenv("MBELT_FILECOIN_STREAMER_KAFKA"), // "localhost:9092",
	}

	banner := "\nMBELT_FILECOIN_STREAMER_API_URL = " + conf.APIUrl + "\n" +
		"MBELT_FILECOIN_STREAMER_API_TOKEN = " + conf.APIToken + "\n" +
		"MBELT_FILECOIN_STREAMER_KAFKA = " + conf.KafkaHosts + "\n" +
		"MBELT_FILECOIN_STREAMER_MIN_HEIGHT = " + os.Getenv("MBELT_FILECOIN_STREAMER_MIN_HEIGHT") + "\n"

	log.Println(banner)
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

	startHeight := abi.ChainEpoch(0)

	// Temp
	strHeight := os.Getenv("MBELT_FILECOIN_STREAMER_MIN_HEIGHT")
	if strHeight != "" {
		strHeightVal, _ := strconv.ParseInt(strHeight, 10, 64)
		startHeight = abi.ChainEpoch(strHeightVal)
	}

	for height := startHeight; height < syncHeight; height++ {
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
