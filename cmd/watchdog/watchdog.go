package watchdog

import (
	"context"
	"database/sql"
	"github.com/afiskon/promtail-client/promtail"
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

const watchdogLokiJob = "watchdog"

var lastCheckedHeight int = 0
var logger promtail.Client

func Start(conf *config.Config, exitAfterOneCheck bool, timeBetweenChecks int, checkFrom int) {
	var (
		exitCode int
		err error
	)
	defer os.Exit(exitCode)

	logger, err = services.InitLogger(conf.LokiUrl, conf.LokiSourceName, watchdogLokiJob)
	if err != nil {
		log.Println("[Watchdog][Debug]", "Cannot init promtail client:", err)
		exitCode = 1
		return
	}

	err = services.InitServices(conf)
	if err != nil {
		logger.Errorf("[Watchdog][Debug] Cannot init services: %s", err)
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
	heightFromDb, err := services.App().PgDatastore().GetMaxHeight()
	if err != nil {
		logger.Errorf("[Watchdog][checkConsistency][Error] Can't get max height from postgres DB. Error: %s", err)
		return
	}

	for lastCheckedHeight <= heightFromDb {
		select {
		case <-ctx.Done():
			return
		default:
			// TODO: change for prom metrics
			log.Println("Checking consistency... Last checked height:", lastCheckedHeight, "height from DB:", heightFromDb)

			blocks, state, err := services.App().PgDatastore().GetTipSetBlocksAndStateByHeight(lastCheckedHeight)
			if err != nil && err != sql.ErrNoRows {
				logger.Errorf("[Watchdog][checkConsistency][Error] Can't get tipset's blocks and state from postgres DB. Error: %s", err)
				return
			}

			//ts, _ := services.App().TipSetsService().GetByHeight(abi.ChainEpoch(lastCheckedHeight))

			var cidsEqual bool
			if blocks != nil {
				// TODO

			}

			if err == sql.ErrNoRows || blocks != nil || !cidsEqual {
				services.App().TipSetsService().PushTipSetsToRevert(lastCheckedHeight, ctx)
				ts, _ := services.App().SyncService().SyncTipSetForHeight(abi.ChainEpoch(lastCheckedHeight))
				services.App().SyncService().CollectAndPushOtherEntitiesByTipSet(ts, ctx)
				lastCheckedHeight++
				continue
			}

			if state != 0 {
				lastCheckedHeight++
				continue
			}




			//TODO: what if we don't even have tipset of this height in db

			countBlocks, err := services.App().PgDatastore().GetBlocksCountByHeight(lastCheckedHeight)
			if err != nil {
				logger.Errorf("[Watchdog][checkConsistency][Error] Can't get blocks count by height from postgres DB. Error: %s", err)
				return
			}

			countMsgs, err := services.App().PgDatastore().GetMessagesCountByHeight(lastCheckedHeight)
			if err != nil {
				logger.Errorf("[Watchdog][checkConsistency][Error] Can't get messages count by height from postgres DB. Error: %s")
				return
			}

			msgCids := make(map[cid.Cid]struct{})
			for _, cidRaw := range blocks {
				blkCid, err := cid.Decode(cidRaw)
				if err != nil {
					logger.Errorf("[Watchdog][checkConsistency][Error] Couldn't decode block cid. Error: %s", err)
				}

				blkMsgs := services.App().MessagesService().GetBlockMessages(blkCid)
				for _, msgCid := range blkMsgs.Cids {
					msgCids[msgCid] = struct{}{}
				}
			}

			if len(blocks) > countBlocks || len(msgCids) > countMsgs {
				logger.Warnf("Inconsistency found at height: %d. Blocks in DB: %d. Blocks in tipset: %d." +
					" Messages id DB: %d. Messages from lotus: %d. Collecting entities...",
					lastCheckedHeight, countBlocks, len(blocks), countMsgs, len(msgCids))
				ts, _ := services.App().SyncService().SyncTipSetForHeight(abi.ChainEpoch(lastCheckedHeight))
				services.App().SyncService().CollectAndPushOtherEntitiesByTipSet(ts, ctx)
				continue
			}

			lastCheckedHeight++
		}
	}
}

//func cidsEqual(ts *types.TipSet, blocks []string) bool {
//	cids := ts.Cids()
//	if len(cids) != len(blocks) {
//		return false
//	}
//
//	for _, cid := range cids {
//		for _, blCid := range blocks {
//			if
//		}
//	}
//}

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

