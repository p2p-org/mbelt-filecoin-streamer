package commands

import (
	"github.com/k0kubun/pp"
	"github.com/p2p-org/mbelt-filecoin-streamer/watchdog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	//// Used for flags.
	//sync             bool
	//syncForce        bool
	//updHead          bool
	//syncFrom         int
	watchdog_verify_height int
	//--verify  --start 888
	start_height int
	//
	//conf *config.Config

	watchDogCMD = &cobra.Command{
		Use:   "watchdog",
		Short: "A watchdog of filecoin's entities that checks consistency of stored blocks and tipsets",
		//		Long: `This app synchronizes with current filecoin state and keeps in sync by subscribing on it's updates.
		//Entities (tipsets, blocks and messages) are being pushed to Kafka. There are also sinks that get
		//those entities from Kafka streams and push them in PostgreSQL DB.'`,
		Run: func(cmd *cobra.Command, args []string) {
			pp.Println("inside watchdog cobra command")
			watchdog.InitWatcher(conf, start_height)
		},
	}
)

func init() {
	// arguments parsing for watchdog command
	watchDogCMD.PersistentFlags().IntVar(&start_height, "start", 0,
		"Specify start height offset from which watcher will verify consistency of blocks in DB")
	viper.BindPFlag("start", watchDogCMD.PersistentFlags().Lookup("start"))

}
