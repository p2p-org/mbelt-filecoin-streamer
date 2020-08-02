package processor

import (
	"encoding/json"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"log"
	"time"
)

type ProcessorService struct {
	config *config.Config
	ds     *datastore.Datastore
	api    *client.APIClient
}

func Init(config *config.Config, ds *datastore.Datastore, apiClient *client.APIClient) (*ProcessorService, error) {

	return &ProcessorService{
		config: config,
		ds:     ds,
		api:    apiClient,
	}, nil
}

func (s *ProcessorService) ProcessStreams() {
	s.processBlocks()
	return
}

func (s *ProcessorService) processBlocks() error {
	messagesTopic, err := s.ds.InitBlocksTopic()
	if err != nil {
		return err
	}

	for {
		time.Sleep(time.Second)

		msg, err := messagesTopic.ReadMessage(-1)
		if err != nil {
			log.Println("[ProcessorService][Error][ProcessBlocks]", "Cannot get message", err)
			continue
		}

		block := &types.BlockHeader{}
		err = json.Unmarshal(msg.Value, block)
		if err != nil {
			log.Println("[ProcessorService][Error][ProcessBlocks]", "Cannot unmarshal message", err)
			continue
		}

	}

}

func (s *ProcessorService) processBlock(block *types.BlockHeader) {
	// calc
}

func (s *ProcessorService) processMessages() {

}
