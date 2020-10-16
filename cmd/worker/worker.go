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

	// current event is current head. We receive it once right after subscription on head updates
	HeadEventCurrent = "current"
	HeadEventApply   = "apply"
	HeadEventRevert  = "revert"
)

var conf *config.Config

func init() {
	conf = &config.Config{
		APIUrl:     os.Getenv("MBELT_FILECOIN_STREAMER_API_URL"),
		APIWsUrl:   os.Getenv("MBELT_FILECOIN_STREAMER_API_WS_URL"),
		APIToken:   os.Getenv("MBELT_FILECOIN_STREAMER_API_TOKEN"),
		KafkaHosts: os.Getenv("MBELT_FILECOIN_STREAMER_KAFKA"), // "localhost:9092",
		PgUrl:      os.Getenv("MBELT_FILECOIN_STREAMER_PG_URL"),
	}

	banner := "\nMBELT_FILECOIN_STREAMER_API_URL = " + conf.APIUrl + "\n" +
		"MBELT_FILECOIN_STREAMER_API_WS_URL = " + conf.APIWsUrl + "\n" +
		"MBELT_FILECOIN_STREAMER_API_TOKEN = " + conf.APIToken + "\n" +
		"MBELT_FILECOIN_STREAMER_KAFKA = " + conf.KafkaHosts + "\n" +
		"MBELT_FILECOIN_STREAMER_PG_URL = " + conf.PgUrl + "\n"

	log.Println(banner)
}

func main() {
	err := services.InitServices(conf)
	if err != nil {
		log.Println("[App][Debug]", "Cannot init services:", err)
		return
	}

	sync := flag.Bool("sync", true, "Turn on sync starting from last block in DB")
	syncForce := flag.Bool("sync-force", false, "Turn on sync starting from genesis block")
	updHead := flag.Bool("sub-head-updates", true, "Turn on subscription on head updates")
	syncFrom := flag.Int("sync-from", -1, "Height to start sync from. Dont provide or provide negative number to sync from max height in DB")
	syncFromDbOffset := flag.Int("sync-from-db-offset", 100, "Specify offset from max height in DB to start sync from (maxHeightInDb - offset)")

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

	if sync != nil && *sync {
		if syncFrom != nil && *syncFrom >= 0 {
			syncToHead(*syncFrom, syncCtx)
		} else {
			heightFromDb, err := services.App().BlocksService().GetMaxHeightFromDB()
			if err != nil {
				log.Println("Can't get max height from postgres DB, stopping...")
				log.Println(err)
				return
			}

			if syncFromDbOffset != nil && heightFromDb < *syncFromDbOffset {
				syncToHead(0, syncCtx)
			} else if syncFromDbOffset != nil {
				syncToHead(heightFromDb-*syncFromDbOffset, syncCtx)
			} else {
				log.Println("sync-from-db-offset is nil, syncing from max height in DB with no offset")
				syncToHead(heightFromDb, syncCtx)
			}
		}
	}

	if updHead != nil && *updHead {
		updateHeads(updCtx)
	}

	log.Println("mbelt-filecoin-streamer gracefully stopped")
}

func updateHeads(ctx context.Context) {
	headUpdatesCtx, cancelHeadUpdates := context.WithCancel(ctx)
	// Buffer is that big for channel to be able to store some head updates while we are syncing till "current" block
	// TODO: This approach is not solid. Think how we can do it better.
	headUpdates := make(chan []*api.HeadChange, 1000)
	services.App().BlocksService().GetHeadUpdates(headUpdatesCtx, &headUpdates)

	for {
		select {
		case update := <-headUpdates:
			for _, hu := range update {
				switch hu.Type {

				case HeadEventCurrent:
					currentHeight := int(hu.Val.Height())
					maxHeightInDb, err := services.App().BlocksService().GetMaxHeightFromDB()
					if err != nil {
						log.Println("[App][Error][updateHeads]", "couldn't get max height from DB. Error:", err)
						cancelHeadUpdates()
						return
					}
					if currentHeight > maxHeightInDb {
						syncTo(maxHeightInDb, currentHeight, ctx)
					}
					// Pushing block and its messages to kafka just in case.
					// TODO: Duplicates should be handled on db's side
					pushBlocksAndTheirMessages(hu.Val.Blocks())

				case HeadEventRevert:
					services.App().BlocksService().PushBlocksToRevert(hu.Val.Blocks())

				case HeadEventApply:
					pushBlocksAndTheirMessages(hu.Val.Blocks())

				default:
					log.Println("[App][Debug][updateHeads]", "yet unknown event encountered:", hu.Type)
					// Pushing just in case
					pushBlocksAndTheirMessages(hu.Val.Blocks())
				}
			}
		case <-ctx.Done():
			cancelHeadUpdates()
			log.Println("[App][Debug][updateHeads]", "unsubscribed from head updates")
			return
		}
	}
}

func syncToHead(from int, ctx context.Context) {
	head := services.App().TipSetsService().GetHead()
	if head != nil {
		syncTo(from, int(head.Height()), ctx)
	} else {
		syncTo(from, 0, ctx)
	}
}

func syncTo(from int, to int, ctx context.Context) {
	syncHeight := abi.ChainEpoch(to)
	if to <= from {
		log.Println("[App][Debug][sync]", "Specified sync height is too small, syncing to default height:", defaultHeight)
		syncHeight = defaultHeight
	}

	defer log.Println("[App][Debug][sync]", "finished sync")
	startHeight := abi.ChainEpoch(from)
	for height := startHeight; height < syncHeight; {
		select {
		default:
			wg := sync.WaitGroup{}
			wg.Add(batchCapacity)

			for workers := 0; workers < batchCapacity; workers++ {

				go func(height abi.ChainEpoch) {
					defer wg.Done()
					_, tipSet, blocks, msgs := syncForHeight(height)
					services.App().TipSetsService().Push(tipSet)
					services.App().BlocksService().PushBlocks(blocks)
					services.App().MessagesService().Push(msgs)

				}(height)

				height++
			}

			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}

func syncForHeight(height abi.ChainEpoch) (isHeightNotReached bool, tipSet *types.TipSet, blocks []*types.BlockHeader, extMessages []*messages.MessageExtended) {
	log.Println("[Datastore][Debug]", "Load height:", height)

	tipSet, isHeightNotReached = services.App().TipSetsService().GetByHeight(height)

	if !isHeightNotReached {
		log.Println("[App][Debug]", "Height reached")
		return
	}

	// Empty TipSet, skipping
	if tipSet == nil {
		return
	}

	blocks = tipSet.Blocks()
	extMessages = getBlockMessages(blocks)

	return
}

func getBlockMessages(blocks []*types.BlockHeader) (msgs []*messages.MessageExtended) {
	for _, block := range blocks {
		blockMessages := services.App().MessagesService().GetBlockMessages(block.Cid())

		if blockMessages == nil || len(blockMessages.BlsMessages) == 0 {
			continue
		}

		for _, blsMessage := range blockMessages.BlsMessages {
			msgs = append(msgs, &messages.MessageExtended{
				BlockCid:  block.Cid(),
				Message:   blsMessage,
				Timestamp: block.Timestamp,
			})
		}

	}

	return msgs
}

func pushBlocksAndTheirMessages(blocks []*types.BlockHeader) {
	msgs := getBlockMessages(blocks)
	services.App().BlocksService().PushBlocks(blocks)
	services.App().MessagesService().Push(msgs)
}
