package services

import (
	"context"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
)

func InitServices(config *config.Config, kafkaCtx context.Context) error {

	kafkaDs, err := datastore.Init(config, kafkaCtx)
	if err != nil {
		return err
	}

	pgDs, err := pg.Init(config)
	if err != nil {
		return err
	}

	apiClient, err := client.Init(config.APIUrl, config.APIWsUrl, config.APIToken)
	if err != nil {
		return err
	}

	return provider.Init(config, kafkaDs, pgDs, apiClient)
}
