package worker

import (
	"context"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	go gracefulStop(syncCancel, updCancel)

	if syncForce {
		services.App().SyncService().SyncToHead(0, syncCtx)
	}

	if sync {
		if syncFrom >= 0 {
			services.App().SyncService().SyncToHead(syncFrom, syncCtx)
		} else {
			heightFromDb, err := services.App().BlocksService().GetMaxHeightFromDB()
			if err != nil {
				log.Println("Can't get max height from postgres DB, stopping...")
				log.Println(err)
				exitCode = 1
				return
			}

			if heightFromDb < syncFromDbOffset {
				services.App().SyncService().SyncToHead(0, syncCtx)
			} else {
				services.App().SyncService().SyncToHead(heightFromDb-syncFromDbOffset, syncCtx)
			}
		}
	}

	if updHead {
		services.App().SyncService().UpdateHeads(updCtx)
	}

	log.Println("mbelt-filecoin-streamer gracefully stopped")
}

func gracefulStop(syncCancel, updCancel context.CancelFunc) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGHUP)

	sig := <-gracefulStop
	log.Printf("Caught sig: %+v", sig)
	log.Println("Wait for graceful shutdown to finish.")
	syncCancel()
	updCancel()
}
