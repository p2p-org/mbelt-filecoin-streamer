package messages

import (
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
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

func (s *MessagesService) Push(message *types.Message) {
	s.ds.Push(*message)
}
