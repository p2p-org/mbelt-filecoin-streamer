package services

import (
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
)

func InitServices(config *config.Config) error {

	dataStore, err := datastore.Init(config)

	if err != nil {
		return err
	}

	apiClient, err := client.Init(config.APIUrl, config.APIToken)
	if err != nil {
		return err
	}

	return provider.Init(config, dataStore, apiClient)
}
