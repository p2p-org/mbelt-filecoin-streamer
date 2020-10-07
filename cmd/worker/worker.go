package main

import (
	"context"
	"flag"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	defaultHeight = 5000
	batchCapacity = 20
)

var conf *config.Config

func init() {
	conf = &config.Config{
		APIUrl:     os.Getenv("MBELT_FILECOIN_STREAMER_API_URL"),
		APIWsUrl:   os.Getenv("MBELT_FILECOIN_STREAMER_API_WS_URL"),
		APIToken:   os.Getenv("MBELT_FILECOIN_STREAMER_API_TOKEN"),
		KafkaHosts: os.Getenv("MBELT_FILECOIN_STREAMER_KAFKA"), // "localhost:9092",
	}

	banner := "\nMBELT_FILECOIN_STREAMER_API_URL = " + conf.APIUrl + "\n" +
		"MBELT_FILECOIN_STREAMER_API_WS_URL = " + conf.APIWsUrl + "\n" +
		"MBELT_FILECOIN_STREAMER_API_TOKEN = " + conf.APIToken + "\n" +
		"MBELT_FILECOIN_STREAMER_KAFKA = " + conf.KafkaHosts + "\n"

	log.Println(banner)
}

func main() {
	err := services.InitServices(conf)
	if err != nil {
		log.Println("[App][Debug]", "Cannot init services:", err)
		return
	}

	syncFromDB := flag.Bool("sync", true, "Turn on sync starting from last block in DB")
	syncForce := flag.Bool("sync-force", false, "Turn on sync starting from genesis block")
	updHead := flag.Bool("sub-head-updates", true, "Turn on subscription on head updates")
	syncFrom := flag.Int("sync-from", -1, "Turn on subscription on head updates")

	syncCtx, syncCancel := context.WithCancel(context.Background())
	updCtx, updCancel := context.WithCancel(context.Background())

	go func() {
		var gracefulStop = make(chan os.Signal)
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		signal.Notify(gracefulStop, syscall.SIGHUP)

		sig := <-gracefulStop
		log.Printf("Caught sig: %+v", sig)
		log.Println("Wait for graceful shutdown to finish.")
		syncCancel()
		updCancel()
	}()

	if syncForce != nil && *syncForce {
		syncToHead(0, syncCtx)
	}

	if syncFromDB != nil && *syncFromDB {
		if syncFrom != nil && *syncFrom >= 0 {
			syncToHead(*syncFrom, syncCtx)
		} else {
			// get last height from Db and start sync
			//syncToHead(heightFromDb, syncCtx)
			// TODO: Temporary!
			syncToHead(0, syncCtx)
		}
	}

	if updHead != nil && *updHead {
		updateHeads(updCtx)
	}

	log.Println("mbelt-filecoin-streamer gracefully stopped")
}

func updateHeads(ctx context.Context) {
	headUpdatesCtx, cancelHeadUpdates := context.WithCancel(ctx)
	headUpdates := make(chan []*api.HeadChange, 10)
	services.App().BlocksService().GetHeadUpdates(headUpdatesCtx, &headUpdates)

	// TODO: handle "apply", "current" and "revert" events
	// TODO: Also we have to check that we've already synced till "current" block, that we receive first, when subscribing to head events
	for {
		select {
		case update := <-headUpdates:
			for _, hu := range update {
				log.Println("[App][Debug]", "Head updates type", hu.Type)
				log.Println("[App][Debug]", "Head updates val", hu.Val.String())
			}
		case <-ctx.Done():
			cancelHeadUpdates()
			log.Println("[App][Debug][updateHeads]", "unsubscribed from head updates")
			return
		}
	}
}

func syncToHead(from int, ctx context.Context) {
	var syncHeight abi.ChainEpoch
	head := services.App().BlocksService().GetHead()

	if head != nil {
		log.Println("[App][Debug]", "Cannot get head with height:", head.Height())
		syncHeight = head.Height()
	} else {
		log.Println("[App][Debug]", "Cannot get header, use default syncHeight:", defaultHeight)
		syncHeight = defaultHeight
	}

	defer log.Println("[App][Debug][syncToHead]", "finished sync")
	startHeight := abi.ChainEpoch(from)
	for height := startHeight; height < syncHeight; {
		select {
		default:
			wg := sync.WaitGroup{}
			wg.Add(batchCapacity)

			for workers := 0; workers < batchCapacity; workers++ {

				go func(height abi.ChainEpoch) {
					defer wg.Done()
					_, blocks, messages := syncForHeight(height)
					services.App().BlocksService().Push(blocks)
					services.App().MessagesService().Push(messages)

				}(height)

				height++
			}

			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}

func syncForHeight(height abi.ChainEpoch) (isHeightNotReached bool, blocks []*types.BlockHeader, extMessages []*messages.MessageExtended) {
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
	for _, block := range blocks {
		blockMessages := services.App().MessagesService().GetBlockMessages(block.Cid())

		if blockMessages == nil || len(blockMessages.BlsMessages) == 0 {
			continue
		}

		for _, blsMessage := range blockMessages.BlsMessages {
			extMessages = append(extMessages, &messages.MessageExtended{
				BlockCid:  block.Cid(),
				Message:   blsMessage,
				Timestamp: block.Timestamp,
			})
		}

	}
	return
}
