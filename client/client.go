package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/ipfs/go-cid"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	protocolVersion = "2.0"
)

type APIClient struct {
	url   string
	wsUrl string
	jwt   string
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

func Init(url, wsUrl, jwt string) (*APIClient, error) {
	c := &APIClient{
		url:   url,
		wsUrl: wsUrl,
		jwt:   jwt,
	}

	testGenesis := c.GetGenesis()

	if testGenesis == nil {
		log.Println("[APIClient][Error][Init] Cannot init api client")
		return nil, errors.New("cannot get genesis")
	}

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

	client := &http.Client{}
	resp, err := client.Do(request)

	if err != nil {
		log.Println("[APIClient][Error][Send]", err)
		time.Sleep(time.Millisecond * 100)
		return errors.New("cannot process request")
	}

	if resp.StatusCode != 200 {
		log.Println("[APIClient][Error][Send] Error status code", resp.StatusCode)
		return errors.New("cannot process request")
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	return json.Unmarshal(data, &dst)
}

// Methods

func (c *APIClient) GetGenesis() *types.TipSet {
	resp := &TipSet{}
	err := c.do(ChainGetGenesis, nil, resp)
	if err != nil {
		log.Println("[API][Error][GetGenesis]", err)
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
		log.Println("[API][Error][GetHead]", err)
		return nil
	}

	if resp.Error != nil {
		log.Println("[API][Error][GetHead]", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetHeadUpdates(ctx context.Context, resChan *chan []*api.HeadChange) {
	jrpcClient, err := NewClient(c.wsUrl)
	if err != nil {
		log.Println("[API][Error][GetHeadUpdates]", err)
		return
	}

	cons := make(chan []byte, 100)
	_, err = jrpcClient.Subscribe(ChainNotify, nil, &cons)

	if err != nil {
		log.Println("[API][Error][GetHeadUpdates]", err)
		return
	}

	for {
		select {
		case val := <-cons:

			upd := &HeadUpdates{}
			err := json.Unmarshal(val, upd)
			if err != nil {
				log.Println("[API][Error][GetHeadUpdates]", "An error occurred while trying to unmarshal head update", err)
			}

			*resChan <- upd.Params.HeadChanges

		case <-ctx.Done():
			return
		}
	}
}

func (c *APIClient) GetBlock(cid cid.Cid) *types.BlockHeader {
	resp := &Block{}
	err := c.do(ChainGetBlock, []interface{}{cid}, resp)
	if err != nil {
		log.Println("[API][Error][GetBlock]", err)
		return nil
	}

	if resp.Error != nil {
		log.Println("[API][Error][GetBlock]", resp.Error.Message)
		return nil
	}
	return resp.Result
}

func (c *APIClient) GetByHeight(height abi.ChainEpoch) (*types.TipSet, bool) {
	resp := &TipSet{}
	err := c.do(ChainGetTipSetByHeight, []interface{}{height, types.EmptyTSK}, resp)
	if err != nil {
		log.Println("[API][Error][GetByHeight]", err)
		return nil, true
	}

	if resp.Error != nil {
		log.Println("[API][Error][GetByHeight]", resp.Error.Message)
		// Height reaching check
		if strings.Contains(resp.Error.Message, "looking for tipset with height greater than start") {
			return nil, false
		}
		return nil, true
	}
	return resp.Result, true
}

func (c *APIClient) GetBlockMessages(cid cid.Cid) *api.BlockMessages {
	resp := &BlockMessages{}
	err := c.do(ChainGetBlockMessages, []interface{}{cid}, resp)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		log.Println("[API][Error][GetBlockMessages]", resp.Error.Message)
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
