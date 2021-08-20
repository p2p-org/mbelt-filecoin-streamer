package messages

import (
	"context"
	"encoding/base64"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/utils"
	"log"
	"math/big"
)

type MessagesService struct {
	config *config.Config
	ds     *datastore.KafkaDatastore
	api    *client.APIClient
}

type MessageExtended struct {
	Cid           cid.Cid
	BlockCids     map[cid.Cid]struct{}
	Height        abi.ChainEpoch
	Message       *types.Message
	FromId        *address.Address
	ToId          *address.Address
	FromType      string
	ToType        string
	MethodName    string
	Timestamp     uint64
	ParentBaseFee abi.TokenAmount
}

type MessageReceiptWithCid struct {
	Cid     cid.Cid
	Receipt *types.MessageReceipt
}

type MessageFromDb struct {
	Cid        string
	Height     int64
	BlockCids  []string
	Method     int
	MethodName string
	From       string
	FromId     string
	FromType   string
	To         string
	ToId       string
	ToType     string
	Value      big.Int
	GasLimit   big.Int
	GasPremium big.Int
	GasFeeCap  big.Int
	GasUsed    big.Int
	BaseFee    big.Int
	ExitCode   int
	BlockTime  int64
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

func (s *MessagesService) GetParentMessages(blockCid cid.Cid) []*client.MessageAndCid {
	return s.api.GetParentMessages(blockCid)
}

func (s *MessagesService) GetParentReceipts(blockCid cid.Cid) []*types.MessageReceipt {
	return s.api.GetParentReceipts(blockCid)
}

func (s *MessagesService) Push(messages []*MessageExtended, ctx context.Context) {
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
		if message != nil {
			m[message.Cid.String()] = serializeMessage(message)
		}
	}

	s.ds.Push(datastore.TopicMessages, m, ctx)
}

func (s *MessagesService) PushReceipts(receipts []*MessageReceiptWithCid, ctx context.Context) {
	// Empty messages has panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[MessagesService][Recover]", "Throw panic", r)
		}
	}()

	if receipts == nil {
		return
	}

	m := map[string]interface{}{}

	for _, receipt := range receipts {
		if receipt != nil {
			m[receipt.Cid.String()] = serializeMessageReceipt(receipt)
		}
	}

	s.ds.Push(datastore.TopicMessageReceipts, m, ctx)
}

func serializeMessage(extMessage *MessageExtended) map[string]interface{} {
	blockCids := make([]cid.Cid, 0, len(extMessage.BlockCids))
	for k, _ := range extMessage.BlockCids {
		blockCids = append(blockCids, k)
	}

	var fromId, toId string
	if extMessage.FromId != nil {
		fromId = extMessage.FromId.String()
	}
	if extMessage.ToId != nil {
		toId = extMessage.ToId.String()
	}

	result := map[string]interface{}{
		"cid":         extMessage.Cid.String(),
		"height":      extMessage.Height,
		"block_cids":  utils.CidsToVarcharArray(blockCids),
		"method":      extMessage.Message.Method,
		"method_name": extMessage.MethodName,
		"from":        extMessage.Message.From.String(),
		"from_id":     fromId,
		"from_type":   extMessage.FromType,
		"to":          extMessage.Message.To.String(),
		"to_id":       toId,
		"to_type":     extMessage.ToType,
		"value":       extMessage.Message.ValueReceived(),
		"gas_limit":   extMessage.Message.GasLimit,
		"gas_fee_cap": extMessage.Message.GasFeeCap,
		"gas_premium": extMessage.Message.GasPremium,
		"base_fee":    extMessage.ParentBaseFee,
		"params":      extMessage.Message.Params,
		"data":        extMessage.Message,
		// "block_time": time.Unix(int64(extMessage.Timestamp), 0).Format(time.RFC3339),
		"block_time":  extMessage.Timestamp,
	}
	return result
}

func serializeMessageReceipt(receipt *MessageReceiptWithCid) map[string]interface{} {
	return map[string]interface{}{
		"cid":       receipt.Cid.String(),
		"gas_used":  receipt.Receipt.GasUsed,
		"exit_code": receipt.Receipt.ExitCode,
		"return":    base64.StdEncoding.EncodeToString(receipt.Receipt.Return),
	}
}
