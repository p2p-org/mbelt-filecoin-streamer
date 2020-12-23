package commands

import (
	"github.com/p2p-org/mbelt-filecoin-streamer/watchdog"
	"github.com/spf13/cobra"
)

var (
	//// Used for flags.
	//sync             bool
	//syncForce        bool
	//updHead          bool
	//syncFrom         int
	watchdog_verify_height int
	//
	//conf *config.Config

	watchDogCMD = &cobra.Command{
		Use:   "mbelt-filecoin-watchdog",
		Short: "A watchdog of filecoin's entities that checks consistency of stored blocks and tipsets",
		//		Long: `This app synchronizes with current filecoin state and keeps in sync by subscribing on it's updates.
		//Entities (tipsets, blocks and messages) are being pushed to Kafka. There are also sinks that get
		//those entities from Kafka streams and push them in PostgreSQL DB.'`,
		Run: func(cmd *cobra.Command, args []string) {
			watchdog.InitWatcher(conf)
		},
	}
)
