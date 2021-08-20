package commands

import (
	"github.com/p2p-org/mbelt-filecoin-streamer/cmd/worker"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"strconv"
)

var (
	// Used for flags.
	sync             bool
	syncForce        bool
	follow           bool
	updHead          bool
	syncFrom         int
	syncFromDbOffset int

	conf *config.Config

	rootCmd = &cobra.Command{
		Use:   "[--sync | --sync-force] [--follow-chain-sync] [--sub-head-updates] [--sync-from=<height>] [--sync-from-db-offset=<offset>]",
		Short: "A streamer of filecoin's entities to PostgreSQL DB through Kafka",
		Long: `This app synchronizes with current filecoin state and keeps in sync by subscribing on it's updates.
Entities (tipsets, blocks and messages) are being pushed to Kafka. There are also sinks that get
those entities from Kafka streams and push them in PostgreSQL DB.`,
		Run: func(cmd *cobra.Command, args []string) {
			worker.Start(conf, sync, syncForce, follow, updHead, syncFrom, syncFromDbOffset)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVarP(&sync, "sync", "s", true,
		"Turn on sync starting from last block in DB")
	rootCmd.PersistentFlags().BoolVarP(&syncForce, "sync-force", "f", false,
		"Turn on sync starting from genesis block")
	rootCmd.PersistentFlags().BoolVarP(&follow, "follow-chain-sync", "U",true,
		"After syncing to current head follow chain sync quering currHeadHeight + 1, not via head updates")
	rootCmd.PersistentFlags().BoolVarP(&updHead, "sub-head-updates", "u", true,
		"Turn on subscription on head updates")
	rootCmd.PersistentFlags().IntVarP(&syncFrom, "sync-from", "F", -1,
		"Height to start sync from. Dont provide or provide negative number to sync from max height in DB")
	rootCmd.PersistentFlags().IntVarP(&syncFromDbOffset, "sync-from-db-offset", "o", 100,
		"Specify offset from max height in DB to start sync from (maxHeightInDb - offset)")

	viper.BindPFlag("sync", rootCmd.PersistentFlags().Lookup("sync"))
	viper.BindPFlag("sync_force", rootCmd.PersistentFlags().Lookup("sync-force"))
	viper.BindPFlag("follow_chain_sync", rootCmd.PersistentFlags().Lookup("follow-chain-sync"))
	viper.BindPFlag("sub_head_updates", rootCmd.PersistentFlags().Lookup("sub-head-updates"))
	viper.BindPFlag("sync_from", rootCmd.PersistentFlags().Lookup("sync-from"))
	viper.BindPFlag("sync_from_db_offset", rootCmd.PersistentFlags().Lookup("sync-from-db-offset"))
	viper.SetDefault("sync", true)
	viper.SetDefault("sync_force", false)
	viper.SetDefault("follow_chain_sync", true)
	viper.SetDefault("sub_head_updates", true)
	viper.SetDefault("sync_from", -1)
	viper.SetDefault("sync_from_db_offset", 100)
	rootCmd.AddCommand(watchdogCmd)
}

func initConfig() {
	viper.SetEnvPrefix("MBELT_FILECOIN_STREAMER")
	viper.AutomaticEnv()

	conf = &config.Config{
		APIUrl:          viper.GetString("API_URL"),
		APIWsUrl:        viper.GetString("API_WS_URL"),
		APIToken:        viper.GetString("API_TOKEN"),
		KafkaPrefix:     viper.GetString("KAFKA_PREFIX"),
		KafkaHosts:      viper.GetString("KAFKA"), // "localhost:9092",
		KafkaAsyncWrite: viper.GetBool("KAFKA_ASYNC_WRITE"),
		PgUrl:           viper.GetString("PG_URL"),
	}

	banner := "\nMBELT_FILECOIN_STREAMER_API_URL = " + conf.APIUrl + "\n" +
		"MBELT_FILECOIN_STREAMER_API_WS_URL = " + conf.APIWsUrl + "\n" +
		"MBELT_FILECOIN_STREAMER_API_TOKEN = " + conf.APIToken + "\n" +
		"MBELT_FILECOIN_STREAMER_KAFKA = " + conf.KafkaHosts + "\n" +
		"MBELT_FILECOIN_STREAMER_KAFKA_PREFIX = " + conf.KafkaPrefix + "\n" +
		"MBELT_FILECOIN_STREAMER_KAFKA_ASYNC_WRITE = " + strconv.FormatBool(conf.KafkaAsyncWrite) + "\n" +
		"MBELT_FILECOIN_STREAMER_PG_URL = " + conf.PgUrl + "\n"

	log.Println(banner)
}

func Execute() error {
	return rootCmd.Execute()
}