package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/v3/actors/builtin/miner"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/services"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	protocolVersion    = "2.0"
	httpRequestTimeout = 60 * time.Second
	apiClientLokiJob   = "api_client"
)

type APIClient struct {
	url   string
	wsUrl string
	jwt   string

	wsClientPool WsClientPool

	logger promtail.Client
}

type APIErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIRequest struct {
	Id      int           `json:"id"`
	Version string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type APIResponse struct {
	Id      int     `json:"id"`
	Version string  `json:"jsonrpc"` // 2.0
	Error   *APIErr `json:"error"`
}

func Init(conf *config.Config) (*APIClient, error) {
	c := &APIClient{
		url:   conf.APIUrl,
		wsUrl: conf.APIWsUrl,
		jwt:   conf.APIToken,

		wsClientPool: NewWsClientPool(conf.APIWsUrl),
	}

	testGenesis := c.GetGenesis()

	if testGenesis == nil {
		log.Println("Cannot init api client")
		return nil, errors.New("cannot get genesis")
	}

	logger, err := services.InitLogger(conf.LokiUrl, conf.LokiSourceName, apiClientLokiJob)
	if err != nil {
		return nil, err
	}

	c.logger = logger

	return c, nil
}

func (c *APIClient) do(method string, params []interface{}, dst interface{}) error {
	var err error
	payload := APIRequest{
		Id:      time.Now().Nanosecond(),
		Version: protocolVersion,
		Method:  method,
		Params:  params,
	}

	encodedMessage, err := json.Marshal(payload)
	if err != nil {
		return errors.New("cannot marshal request")
	}

	request, err := http.NewRequest("POST", c.url, bytes.NewBuffer(encodedMessage))

	if err != nil {
		return errors.New("cannot create request")
	}

	request.Header.Set("Content-Type", "application/json")
	if c.jwt != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.jwt))
	}

	client := &http.Client{Timeout: httpRequestTimeout}
	resp, err := client.Do(request)

	if err != nil {
		c.logger.Errorf("[APIClient][Error][Send] %s", err)
		time.Sleep(time.Millisecond * 100)
		return errors.New("cannot process request")
	}

	if resp.StatusCode != 200 {
		c.logger.Errorf("[APIClient][Error][Send] Error status code: %d", resp.StatusCode)
		return errors.New("cannot process request")
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	return json.Unmarshal(data, &dst)
}

// Methods

func (c *APIClient) GetGenesis() *types.TipSet {
	client := c.wsClientPool.Get()
	resp := &TipSet{}
	err := client.Do(ChainGetGenesis, nil, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetGenesis] %s", err)
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetGenesis] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetGenesisHttp() *types.TipSet {
	resp := &TipSet{}
	err := c.do(ChainGetGenesis, nil, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetGenesis] %s", err)
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetGenesis] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetHead() *types.TipSet {
	client := c.wsClientPool.Get()
	resp := &TipSet{}
	err := client.Do(ChainHead, nil, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetHead] %s", err)
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetHead] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetHeadHttp() *types.TipSet {
	resp := &TipSet{}
	err := c.do(ChainHead, nil, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetHead] %s", err)
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetHead] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetHeadUpdates(ctx context.Context, resChan *chan []*api.HeadChange) {
	// Creating new client because there is no profit to get it from pool because we will never put it back
	jrpcClient, err := NewClient(c.wsUrl)
	if err != nil {
		c.logger.Errorf("[API][Error][GetHeadUpdates] %s", err)
		return
	}

	cons := make(chan []byte, 100)
	subCtx, subCancel := context.WithCancel(ctx)
	_, err = jrpcClient.Subscribe(ChainNotify, nil, &cons, subCtx)
	if err != nil {
		c.logger.Errorf("[API][Error][GetHeadUpdates] %s", err)
		return
	}

	for {
		select {
		case val := <-cons:
			upd := &HeadUpdates{}
			err := json.Unmarshal(val, upd)
			if err != nil {
				c.logger.Errorf("[API][Error][GetHeadUpdates] An error occurred while trying to unmarshal head update. Error: %s", err)
			}

			*resChan <- upd.Params.HeadChanges

		case <-ctx.Done():
			subCancel()
			return
		}
	}
}

func (c *APIClient) GetBlock(cid cid.Cid) *types.BlockHeader {
	client := c.wsClientPool.Get()
	resp := &Block{}
	err := client.Do(ChainGetBlock, []interface{}{cid}, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetBlock] %s", err)
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetBlock] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetBlockHttp(cid cid.Cid) *types.BlockHeader {
	resp := &Block{}
	err := c.do(ChainGetBlock, []interface{}{cid}, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetBlock] %s", err)
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetBlock] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetByHeight(height abi.ChainEpoch) (*types.TipSet, bool) {
	client := c.wsClientPool.Get()
	resp := &TipSet{}
	err := client.Do(ChainGetTipSetByHeight, []interface{}{height, types.EmptyTSK}, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetByHeight] %s", err)
		return nil, false
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		// Height reaching check
		if strings.Contains(resp.Error.Message, "looking for tipset with height greater than start") {
			return nil, false
		}
		c.logger.Errorf("[API][Error][GetByHeight] %s", resp.Error.Message)
		return nil, true
	}
	return resp.Result, false
}

func (c *APIClient) GetByHeightHttp(height abi.ChainEpoch) (*types.TipSet, bool) {
	resp := &TipSet{}
	err := c.do(ChainGetTipSetByHeight, []interface{}{height, types.EmptyTSK}, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetByHeight] %s", err)
		return nil, true
	}

	if resp.Error != nil {
		// Height reaching check
		if strings.Contains(resp.Error.Message, "looking for tipset with height greater than start") {
			return nil, false
		}
		c.logger.Errorf("[API][Error][GetByHeight] %s", resp.Error.Message)
		return nil, true
	}
	return resp.Result, true
}

func (c *APIClient) GetByKey(key types.TipSetKey) *types.TipSet {
	client := c.wsClientPool.Get()
	resp := &TipSet{}
	err := client.Do(ChainGetTipSet, []interface{}{key}, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetByKey] %s", err)
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetByKey] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetByKeyHttp(key types.TipSetKey) *types.TipSet {
	resp := &TipSet{}
	err := c.do(ChainGetTipSet, []interface{}{key}, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][GetByKey] %s", err)
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetByKey] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetBlockMessages(cid cid.Cid) *api.BlockMessages {
	client := c.wsClientPool.Get()
	resp := &BlockMessages{}
	err := client.Do(ChainGetBlockMessages, []interface{}{cid}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetBlockMessages] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetBlockMessagesHttp(cid cid.Cid) *api.BlockMessages {
	resp := &BlockMessages{}
	err := c.do(ChainGetBlockMessages, []interface{}{cid}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetBlockMessages] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMessage(cid cid.Cid) *types.Message {
	client := c.wsClientPool.Get()
	resp := &Message{}
	err := client.Do(ChainGetMessage, []interface{}{cid}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetMessage] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMessageHttp(cid cid.Cid) *types.Message {
	resp := &Message{}
	err := c.do(ChainGetMessage, []interface{}{cid}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetMessage] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetParentMessages(blockCid cid.Cid) []*MessageAndCid {
	client := c.wsClientPool.Get()
	resp := &MessageAndCidResponse{}
	err := client.Do(ChainGetParentMessages, []interface{}{blockCid}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetMessage] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetParentMessagesHttp(blockCid cid.Cid) []*MessageAndCid {
	resp := &MessageAndCidResponse{}
	err := c.do(ChainGetParentMessages, []interface{}{blockCid}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetMessage] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetParentReceipts(blockCid cid.Cid) []*types.MessageReceipt {
	client := c.wsClientPool.Get()
	resp := &MessageReceiptResponse{}
	err := client.Do(ChainGetParentReceipts, []interface{}{blockCid}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetMessage] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetParentReceiptsHttp(blockCid cid.Cid) []*types.MessageReceipt {
	resp := &MessageReceiptResponse{}
	err := c.do(ChainGetParentReceipts, []interface{}{blockCid}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetMessage] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) ChainHasObj(cid cid.Cid) (bool, error) {
	client := c.wsClientPool.Get()
	resp := &HasObj{}
	err := client.Do(ChainHasObj, []interface{}{cid}, resp)
	if err != nil {
		return false, err
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ChainHasObj] %s", resp.Error.Message)
		return false, errors.New(resp.Error.Message)
	}
	return resp.Result, nil
}

func (c *APIClient) ChainHasObjHttp(cid cid.Cid) (bool, error) {
	resp := &HasObj{}
	err := c.do(ChainHasObj, []interface{}{cid}, resp)
	if err != nil {
		return false, err
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ChainHasObj] %s", resp.Error.Message)
		return false, errors.New(resp.Error.Message)
	}
	return resp.Result, nil
}

func (c *APIClient) GetActor(addr address.Address, tsk *types.TipSetKey) *types.Actor {
	client := c.wsClientPool.Get()
	resp := &Actor{}
	var err error
	if tsk == nil {
		err = client.Do(StateGetActor, []interface{}{addr, nil}, resp)
	} else {
		err = client.Do(StateGetActor, []interface{}{addr, *tsk}, resp)
	}
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetActor] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetActorHttp(addr address.Address, tsk *types.TipSetKey) *types.Actor {
	resp := &Actor{}
	var err error
	if tsk == nil {
		err = c.do(StateChangedActors, []interface{}{addr, nil}, resp)
	} else {
		err = c.do(StateChangedActors, []interface{}{addr, *tsk}, resp)
	}
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetActor] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetChangedActors(start, end cid.Cid) map[string]types.Actor {
	client := c.wsClientPool.Get()
	resp := &Actors{}
	err := client.Do(StateChangedActors, []interface{}{start, end}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetChangedActors] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetChangedActorsHttp(start, end cid.Cid) map[string]types.Actor {
	resp := &Actors{}
	err := c.do(StateChangedActors, []interface{}{start, end}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][GetChangedActors] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) ReadState(actor address.Address, tsk types.TipSetKey) *ActorState {
	client := c.wsClientPool.Get()
	resp := &ActorStateResponse{}
	err := client.Do(StateReadState, []interface{}{actor, tsk}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ReadState] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) ReadStateHttp(actor address.Address, tsk types.TipSetKey) *ActorState {
	resp := &ActorStateResponse{}
	err := c.do(StateReadState, []interface{}{actor, tsk}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ReadState] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) ListMiners(tsk types.TipSetKey) []address.Address {
	client := c.wsClientPool.Get()
	resp := &AddressListResponse{}
	err := client.Do(StateListMiners, []interface{}{tsk}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ListMiners] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) ListMinersHttp(tsk types.TipSetKey) []address.Address {
	resp := &AddressListResponse{}
	err := c.do(StateListMiners, []interface{}{tsk}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ListMiners] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMinerInfo(actor address.Address, tsk types.TipSetKey) *miner.MinerInfo {
	client := c.wsClientPool.Get()
	resp := &MinerInfoResponse{}
	err := client.Do(StateMinerInfo, []interface{}{actor, tsk}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ListMiners] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMinerInfoHttp(actor address.Address, tsk types.TipSetKey) *miner.MinerInfo {
	resp := &MinerInfoResponse{}
	err := c.do(StateMinerInfo, []interface{}{actor, tsk}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ListMiners] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMinerPower(actor address.Address, tsk types.TipSetKey) *api.MinerPower {
	client := c.wsClientPool.Get()
	resp := &MinerPowerResponse{}
	err := client.Do(StateMinerPower, []interface{}{actor, tsk}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ListMiners] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMinerPowerHttp(actor address.Address, tsk types.TipSetKey) *api.MinerPower {
	resp := &MinerPowerResponse{}
	err := c.do(StateMinerPower, []interface{}{actor, tsk}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ListMiners] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMinerSectors(actor address.Address, tsk types.TipSetKey) []*miner.SectorOnChainInfo {
	client := c.wsClientPool.Get()
	resp := &MinerSectorsResponse{}
	err := client.Do(StateMinerSectors, []interface{}{actor, nil, tsk}, resp)
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ListMiners] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMinerSectorsHttp(actor address.Address, tsk types.TipSetKey) []*miner.SectorOnChainInfo {
	resp := &MinerSectorsResponse{}
	err := c.do(StateMinerSectors, []interface{}{actor, nil, tsk}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][ListMiners] %s", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) LookupID(actor address.Address, tsk *types.TipSetKey) *address.Address {
	client := c.wsClientPool.Get()
	resp := &AddressResponse{}
	var err error
	if tsk == nil {
		err = client.Do(StateLookupID, []interface{}{actor, nil}, resp)
	} else {
		err = client.Do(StateLookupID, []interface{}{actor, *tsk}, resp)
	}
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][LookupID] %s", resp.Error.Message)
		return nil
	}
	return &resp.Result
}

func (c *APIClient) LookupIDHttp(actor address.Address, tsk *types.TipSetKey) *address.Address {
	resp := &AddressResponse{}
	var err error
	if tsk == nil {
		err = c.do(StateLookupID, []interface{}{actor, nil}, resp)
	} else {
		err = c.do(StateLookupID, []interface{}{actor, *tsk}, resp)
	}
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][LookupID] %s", resp.Error.Message)
		return nil
	}
	return &resp.Result
}

func (c *APIClient) AccountKey(actor address.Address, tsk *types.TipSetKey) *address.Address {
	client := c.wsClientPool.Get()
	resp := &AddressResponse{}
	var err error
	if tsk == nil {
		err = client.Do(StateAccountKey, []interface{}{actor, nil}, resp)
	} else {
		err = client.Do(StateAccountKey, []interface{}{actor, *tsk}, resp)
	}
	if err != nil {
		return nil
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][AccountKey] %s", resp.Error.Message)
		return nil
	}
	return &resp.Result
}

func (c *APIClient) AccountKeyHttp(actor address.Address, tsk *types.TipSetKey) *address.Address {
	resp := &AddressResponse{}
	var err error
	if tsk == nil {
		err = c.do(StateAccountKey, []interface{}{actor, nil}, resp)
	} else {
		err = c.do(StateAccountKey, []interface{}{actor, *tsk}, resp)
	}
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][AccountKey] %s", resp.Error.Message)
		return nil
	}
	return &resp.Result
}

func (c *APIClient) NetworkName() string {
	client := c.wsClientPool.Get()
	resp := &StringResponse{}

	err := client.Do(StateNetworkName, []interface{}{}, resp)

	if err != nil {
		c.logger.Errorf("[API][Error][NetworkName] %s", err)
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][NetworkName] %s", resp.Error.Message)
	}

	return resp.Result
}

func (c *APIClient) NetworkNameHttp() string {
	resp := &StringResponse{}

	err := c.do(StateNetworkName, []interface{}{}, resp)

	if err != nil {
		c.logger.Errorf("[API][Error][NetworkName] %s", err)
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][NetworkName] %s", resp.Error.Message)
	}
	return resp.Result
}

func (c *APIClient) NetworkVersion(tsk *types.TipSetKey) int {
	client := c.wsClientPool.Get()
	resp := &IntResponse{}

	var err error
	if tsk == nil {
		err = c.do(StateNetworkVersion, []interface{}{nil}, resp)
	} else {
		err = c.do(StateNetworkVersion, []interface{}{*tsk}, resp)
	}

	if err != nil {
		c.logger.Errorf("[API][Error][NetworkVersion] %s", err)
	}

	c.wsClientPool.Put(client)

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][NetworkVersion] %s", resp.Error.Message)
	}

	return resp.Result
}

func (c *APIClient) NetworkVersionHttp(tsk *types.TipSetKey) int {
	resp := &IntResponse{}

	var err error
	if tsk == nil {
		err = c.do(StateNetworkName, []interface{}{nil}, resp)
	} else {
		err = c.do(StateNetworkName, []interface{}{*tsk}, resp)
	}

	if err != nil {
		c.logger.Errorf("[API][Error][NetworkVersion] %s", err)
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][NetworkVersion] %s", resp.Error.Message)
	}
	return resp.Result
}

func (c *APIClient) NetPeersHttp() []*peer.AddrInfo {
	resp := &PeersResponse{}

	err := c.do(NetPeers, []interface{}{}, resp)
	if err != nil {
		c.logger.Errorf("[API][Error][NetPeers] %s", err)
	}

	if resp.Error != nil {
		c.logger.Errorf("[API][Error][NetPeers] %s", resp.Error.Message)
	}
	return resp.Result
}