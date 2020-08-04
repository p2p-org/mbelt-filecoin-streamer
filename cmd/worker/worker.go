package main

import (
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"log"
	"os"
	"strconv"
	"sync"
)

const (
	defaultHeight = 5000
	batchCapacity = 20
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

	if head != nil {
		log.Println("[App][Debug]", "Cannot got head with height:", head.Height())
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

	for height := startHeight; height < syncHeight; {

		wg := sync.WaitGroup{}
		wg.Add(batchCapacity)

		for workers := 0; workers < batchCapacity; workers++ {

			go func(height abi.ChainEpoch) {
				defer wg.Done()
				_, blocks, messages := syncBlocks(height)
				services.App().BlocksService().Push(blocks)
				services.App().MessagesService().Push(messages)

			}(height)

			height++
		}

		wg.Wait()
	}
}

func syncBlocks(height abi.ChainEpoch) (isHeightNotReached bool, blocks []*types.BlockHeader, messages []*types.Message) {
	log.Println("[Datastore][Debug]", "Load height:", height)

	tipSet, isHeightNotReached := services.App().BlocksService().GetByHeight(height)

	if !isHeightNotReached {
		log.Println("[App][Debug]", "Height reached")
		return
	}

	// Empty TipSet, skipping
	if tipSet == nil {
		return
	}

	blocks = tipSet.Blocks()

	for _, block := range tipSet.Blocks() {
		if block.Messages.Defined() {
			blockMessages := services.App().MessagesService().GetBlockMessages(block.Messages)

			if blockMessages == nil {
				continue
			}

			if len(blockMessages.Cids) > 0 {
				for _, messageCid := range blockMessages.Cids {
					message := services.App().MessagesService().GetMessage(messageCid)

					if message == nil {
						continue
					}

					messages = append(messages, message)
				}
			}
		}
	}
	return
}
