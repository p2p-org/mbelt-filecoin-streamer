package commands

import (
	"github.com/p2p-org/mbelt-filecoin-streamer/cmd/watchdog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	exitAfterOneCheck bool
	timeBetweenChecks int
	checkFrom         int

	watchdogCmd = &cobra.Command{
		Use:   "watchdog [--exit-after-one-check] [--time-between-checks=<seconds>] [--check-from=<height>]",
		Short: "Mbelt filecoin streamer's database consistency checker.",
		Long: `Watchdog checks database filled by mbelt filecoin streamer for missed entities and collects them if there are any.`,
		Run: func(cmd *cobra.Command, args []string) {
			watchdog.Start(conf, exitAfterOneCheck, timeBetweenChecks, checkFrom)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	watchdogCmd.PersistentFlags().BoolVarP(&exitAfterOneCheck, "exit-after-one-check", "e", false,
		"Don's start checks by timer, just exit after one check")
	watchdogCmd.PersistentFlags().IntVarP(&timeBetweenChecks, "time-between-checks", "t", 35,
		"Time to wait before starting new check.")
	watchdogCmd.PersistentFlags().IntVarP(&checkFrom, "check-from", "c", 0,
		"Height to start checks from")

	viper.BindPFlag("exit_after_one_check", watchdogCmd.PersistentFlags().Lookup("exit-after-one-check"))
	viper.BindPFlag("time_between_checks", watchdogCmd.PersistentFlags().Lookup("time-between-checks"))
	viper.BindPFlag("check_from", watchdogCmd.PersistentFlags().Lookup("check-from"))
	viper.SetDefault("exit_after_one_check", false)
	viper.SetDefault("time_between_checks", 35)
	viper.SetDefault("check_from", 0)
}