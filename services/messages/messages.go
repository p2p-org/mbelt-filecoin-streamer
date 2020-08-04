package messages

import (
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	log "log"
)

type MessagesService struct {
	config *config.Config
	ds     *datastore.Datastore
	api    *client.APIClient
}

func Init(config *config.Config, ds *datastore.Datastore, apiClient *client.APIClient) (*MessagesService, error) {
	return &MessagesService{
		config: config,
		ds:     ds,
		api:    apiClient,
	}, nil
}

func (s *MessagesService) GetBlockMessages(cid cid.Cid) *api.BlockMessages {
	return s.api.GetBlockMessages(cid)
}

func (s *MessagesService) GetMessage(cid cid.Cid) *types.Message {
	return s.api.GetMessage(cid)
}

func (s *MessagesService) Push(messages []*types.Message) {
	// Empty messages has panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[MessagesService][Recover]", "Cid throw panic")
		}
	}()

	if len(messages) == 0 {
		return
	}

	m := map[string]interface{}{}

	for _, message := range messages {
		m[message.Cid().String()] = message
	}

	s.ds.Push(datastore.TopicMessages, m)
}
