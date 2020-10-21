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
	ds     *datastore.KafkaDatastore
	api    *client.APIClient
}

type MessageExtended struct {
	BlockCid  cid.Cid
	Message   *types.Message
	Timestamp uint64
}

func Init(config *config.Config, ds *datastore.KafkaDatastore, apiClient *client.APIClient) (*MessagesService, error) {
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

func (s *MessagesService) Push(messages []*MessageExtended) {
	// Empty messages has panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[MessagesService][Recover]", "Throw panic", r)
		}
	}()

	if messages == nil {
		return
	}

	m := map[string]interface{}{}

	for _, message := range messages {
		m[message.Message.Cid().String()] = serializeMessage(message)
	}

	s.ds.Push(datastore.TopicMessages, m)
}

func serializeMessage(extMessage *MessageExtended) map[string]interface{} {

	result := map[string]interface{}{
		"cid":       extMessage.Message.Cid().String(),
		"block_cid": extMessage.BlockCid.String(),
		"method":    extMessage.Message.Method,
		"from":      extMessage.Message.From.String(),
		"to":        extMessage.Message.To.String(),
		"value":     extMessage.Message.ValueReceived(),
		"gas": map[string]interface{}{
			"limit":   extMessage.Message.GasLimit,
			"fee_cap": extMessage.Message.GasFeeCap,
			"premium": extMessage.Message.GasPremium,
		},
		"params": extMessage.Message.Params,
		"data":   extMessage.Message,
		// "block_time": time.Unix(int64(extMessage.Timestamp), 0).Format(time.RFC3339),
		"block_time": extMessage.Timestamp,
	}
	return result
}
