package client

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	jsonRPCVersion   = "2.0"
	wsRequestTimeout = 2 * time.Second
)

type RPCClient struct {
	conn *websocket.Conn
}

type Request struct {
	Id      uint64        `json:"id"`
	Version string        `json:"jsonrpc"` // 2.0
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type SubResponse struct {
	Version string `json:"jsonrpc"` // 2.0
	Result  int    `json:"result"`
	Id      int    `json:"id"`
}

func NewClient(url string) (c *RPCClient, err error) {
	c = &RPCClient{}
	c.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("[RPCClient][Error][NewClient]", err)
		return nil, err
	}

	return c, nil
}

func (c *RPCClient) readLoop(consumer *chan []byte, ctx context.Context) {
	for {
		select {
		default:
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				log.Println("[RPCClient][Error][readLoop]", err)
				continue
			}

			*consumer <- msg

		case <-ctx.Done():
			err := c.conn.Close()
			if err != nil {
				log.Println("[RPCClient][Error][readLoop]", "Couldn't close ws connection", err)
			}
			return
		}
	}
}

func (c *RPCClient) Subscribe(method string, params []interface{}, consumer *chan []byte, ctx context.Context) (int, error) {
	if consumer == nil {
		return -1, errors.New("consumer channel can't be nil")
	}

	request := &Request{
		Id:      0,
		Version: jsonRPCVersion,
		Method:  method,
		Params:  params,
	}

	err := c.conn.WriteJSON(request)
	if err != nil {
		log.Println("[RPCClient][Error][Subscribe]", err)
		return -1, err
	}

	subMsgCtx, subMsgCancel := context.WithTimeout(ctx, wsRequestTimeout)
	defer subMsgCancel()

	subChan := make(chan SubResponse)

	// Receiving first message after subscription, it should be SubResponse type.
	// Goroutine and select from channel are for timeout only.
	go func() {
		sub := &SubResponse{}
		err := c.conn.ReadJSON(sub)
		if err != nil {
			log.Println("[RPCClient][Error][Subscribe]", err)
		}
		subChan <- *sub
	}()

	sub := SubResponse{}
	select {
	case sub = <-subChan:
		// Do nothing
	case <-subMsgCtx.Done():
		return -1, err
	}

	// Other messages handled asynchronously and sent to consumer chan.
	go c.readLoop(consumer, ctx)

	return sub.Id, nil
}
