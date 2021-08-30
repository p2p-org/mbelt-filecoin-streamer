package services

import (
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
)

func InitServices(config *config.Config) error {

	kafkaDs, err := datastore.Init(config)
	if err != nil {
		return err
	}

	pgDs, err := pg.Init(config)
	if err != nil {
		return err
	}

	apiClient, err := client.Init(config)
	if err != nil {
		return err
	}

	return provider.Init(config, kafkaDs, pgDs, apiClient)
}