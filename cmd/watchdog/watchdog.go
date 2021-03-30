package watchdog

import (
	"context"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var lastCheckedHeight int = 0

func Start(conf *config.Config, exitAfterOneCheck bool, timeBetweenChecks int, checkFrom int) {
	exitCode := 0
	defer os.Exit(exitCode)

	err := services.InitServices(conf)
	if err != nil {
		log.Println("[Watchdog][Debug]", "Cannot init services:", err)
		exitCode = 1
		return
	}

	lastCheckedHeight = checkFrom

	ctx, cancel := context.WithCancel(context.Background())

	go gracefulStop(cancel)

	checkConsistency(ctx)

	if exitAfterOneCheck {
		return
	}

	scheduleChecks(timeBetweenChecks, ctx)

	log.Println("mbelt-filecoin-streamer watchdog gracefully stopped")
}

func checkConsistency(ctx context.Context) {
	heightFromDb, err := services.App().BlocksService().GetMaxHeightFromDB()
	if err != nil {
		log.Println("[Watchdog][checkConsistency][Error] Can't get max height from postgres DB...")
		log.Println(err)
		return
	}

	for lastCheckedHeight <= heightFromDb {
		select {
		case <-ctx.Done():
			return
		default:
			log.Println("Checking consistency... Last checked height:", lastCheckedHeight, "height from DB:", heightFromDb)

			blocks, state, err := services.App().PgDatastore().GetTipSetBlocksAndStateByHeight(lastCheckedHeight)
			if err != nil {
				log.Println("[Watchdog][checkConsistency][Error] Can't get tipset's blocks and state from postgres DB...")
				log.Println(err)
				return
			}

			//TODO: what if we don't even have tipset of this height in db

			if state != 0 {
				lastCheckedHeight++
				continue
			}

			countBlocks, err := services.App().PgDatastore().GetBlocksCountByHeight(lastCheckedHeight)
			if err != nil {
				log.Println("[Watchdog][checkConsistency][Error] Can't get blocks count by height from postgres DB...")
				log.Println(err)
				return
			}

			countMsgs, err := services.App().PgDatastore().GetMessagesCountByHeight(lastCheckedHeight)
			if err != nil {
				log.Println("[Watchdog][checkConsistency][Error] Can't get messages count by height from postgres DB...")
				log.Println(err)
				return
			}

			msgCids := make(map[cid.Cid]struct{})
			for _, cidRaw := range blocks {
				blkCid, err := cid.Decode(cidRaw)
				if err != nil {
					log.Println("[Watchdog][checkConsistency][Error] Couldn't decode block cid.")
					log.Println(err)
				}

				blkMsgs := services.App().MessagesService().GetBlockMessages(blkCid)
				for _, msgCid := range blkMsgs.Cids {
					msgCids[msgCid] = struct{}{}
				}
			}

			if len(blocks) > countBlocks || len(msgCids) > countMsgs {
				log.Println("Inconsistency found at height", lastCheckedHeight, "blocks in DB:", countBlocks,
					"blocks in tipset:", len(blocks), "messages in DB:", countMsgs, "messages from lotus:", len(msgCids),
					"collecting entities")
				ts, _ := services.App().SyncService().SyncTipSetForHeight(abi.ChainEpoch(lastCheckedHeight))
				services.App().SyncService().CollectAndPushOtherEntitiesByTipSet(ts, ctx)
				continue
			}

			lastCheckedHeight++
		}
	}
}

func scheduleChecks(timeBetweenChecks int, ctx context.Context) {
	ticker := time.Tick(time.Duration(timeBetweenChecks) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			checkConsistency(ctx)
		}
	}
}

func gracefulStop(cancel context.CancelFunc) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGHUP)

	sig := <-gracefulStop
	log.Printf("Caught sig: %+v", sig)
	log.Println("Wait for graceful shutdown to finish.")
	cancel()
}

