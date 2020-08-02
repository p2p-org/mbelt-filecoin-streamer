package client

import (
	"errors"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/ipfs/go-cid"
	"golang.org/x/net/websocket"
	"log"
	"time"
)

type APIClient struct {
	conf *websocket.Config
	conn *websocket.Conn
}

type APIErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIRequest struct {
	Id     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type APIResponse struct {
	Id      int     `json:"id"`
	Version string  `json:"jsonrpc"` // 2.0
	Error   *APIErr `json:"error"`
}

func Init(url, origin, jwt string) (*APIClient, error) {
	var (
		c   APIClient
		err error
	)

	c.conf, err = websocket.NewConfig(url, origin)
	if err != nil {
		log.Println("[APIClient][Error][Init]", err)
		return nil, errors.New("cannot init config")
	}

	if jwt != "" {
		c.conf.Header.Add(`Authorization`, `Bearer `+jwt)
	}

	c.conn, err = websocket.DialConfig(c.conf)

	if err != nil {
		log.Println("[APIClient][Error][Init]", err)
		return nil, errors.New("cannot connect")
	}

	return &c, nil
}

func (c *APIClient) do(method string, params []interface{}, dst interface{}) error {
	var err error
	request := APIRequest{
		Id:     1,
		Method: method,
		Params: params,
	}
	for attempt := 3; attempt > 0; attempt-- {
		if err = websocket.JSON.Send(c.conn, request); err != nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		err = websocket.JSON.Receive(c.conn, dst)
		if err != nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		return nil
	}

	if err != nil {
		// Dump response

		var rawResponse string

		if err = websocket.JSON.Send(c.conn, request); err != nil {
			log.Println("[APIClient][Error][Send][Dump]", err)
		}

		err = websocket.Message.Receive(c.conn, &rawResponse)
		if err != nil {
			log.Println("[APIClient][Error][Receive][Dump]", err)
		}

		// log.Println("[APIClient][Error][Receive]", rawResponse)

	}

	return errors.New("receive error")
}

// Methods

func (c *APIClient) GetGenesis() *types.TipSet {
	resp := &TipSet{}
	err := c.do(ChainGetGenesis, nil, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		log.Println("[API][Error][GetGenesis]", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetHead() *types.TipSet {
	resp := &TipSet{}
	err := c.do(ChainHead, nil, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		log.Println("[API][Error][GetHead]", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetBlock(cid cid.Cid) *types.BlockHeader {
	resp := &Block{}
	err := c.do(ChainGetBlock, []interface{}{cid}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		log.Println("[API][Error][GetBlock]", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetByHeight(height abi.ChainEpoch) *types.TipSet {
	resp := &TipSet{}
	err := c.do(ChainGetTipSetByHeight, []interface{}{height, types.EmptyTSK}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		// log.Println("[API][Error][GetByHeight]", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetBlockMessages(cid cid.Cid) *api.BlockMessages {
	resp := &BlockMessages{}
	err := c.do(ChainGetBlockMessages, []interface{}{cid}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		// log.Println("[API][Error][GetBlockMessages]", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetMessage(cid cid.Cid) *types.Message {
	resp := &Message{}
	err := c.do(ChainGetMessage, []interface{}{cid}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		log.Println("[API][Error][GetMessage]", resp.Error.Message)
		return nil
	}
	return resp.Result
}
