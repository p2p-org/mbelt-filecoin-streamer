package worker

import (
	"context"
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

func Start(conf *config.Config, sync bool, syncForce bool, updHead bool, syncFrom int, syncFromDbOffset int) {
	exitCode := 0
	defer os.Exit(exitCode)

	err := services.InitServices(conf)
	if err != nil {
		log.Println("[App][Debug]", "Cannot init services:", err)
		exitCode = 1
		return
	}

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

	if syncForce {
		syncToHead(0, syncCtx)
	}

	if sync {
		if syncFrom >= 0 {
			syncToHead(syncFrom, syncCtx)
		} else {
			heightFromDb, err := services.App().BlocksService().GetMaxHeightFromDB()
			if err != nil {
				log.Println("Can't get max height from postgres DB, stopping...")
				log.Println(err)
				exitCode = 1
				return
			}

			if heightFromDb < syncFromDbOffset {
				syncToHead(0, syncCtx)
			} else {
				syncToHead(heightFromDb-syncFromDbOffset, syncCtx)
			}
		}
	}

	if updHead {
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
					pushTipsetWithBlocksAndMessages(hu.Val)

				case HeadEventRevert:
					services.App().TipSetsService().PushTipSetsToRevert(hu.Val)

				case HeadEventApply:
					pushTipsetWithBlocksAndMessages(hu.Val)

				default:
					log.Println("[App][Debug][updateHeads]", "yet unknown event encountered:", hu.Type)
					// Pushing just in case
					pushTipsetWithBlocksAndMessages(hu.Val)
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
	if startHeight <= 1 {
		log.Println("getting genesis")
		genesis := services.App().TipSetsService().GetGenesis()
		log.Println(genesis.MarshalJSON())
		services.App().TipSetsService().Push(genesis)
		services.App().BlocksService().Push(genesis.Blocks())
		services.App().MessagesService().Push(getBlockMessages(genesis.Blocks()))
	}

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
					services.App().BlocksService().Push(blocks)
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

func pushTipsetWithBlocksAndMessages(tipset *types.TipSet) {
	services.App().TipSetsService().Push(tipset)
	msgs := getBlockMessages(tipset.Blocks())
	services.App().BlocksService().Push(tipset.Blocks())
	services.App().MessagesService().Push(msgs)
}
