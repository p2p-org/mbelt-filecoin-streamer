package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	jsonRPCVersion    = "2.0"
	wsRequestTimeout  = 60 * time.Second
	maxRequestRetries = 5
	maxDialReties     = 10
)

type RPCClient struct {
	conn   *websocket.Conn
	nextId uint64
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
	c = &RPCClient{
		nextId: 0,
	}
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 3 * time.Second,
	}
	c.conn = dialWithRetry(dialer, url, 3)

	return c, nil
}

func dialWithRetry(dialer *websocket.Dialer, url string, retryCount int) (conn *websocket.Conn) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		if retryCount < maxDialReties {
			retryCount++
			log.Println("[RPCCient][Debug][Dial]", "Couldn't dial to lotus ws url. err:", err, "Attempting retry number", retryCount)
			return dialWithRetry(dialer, url, retryCount)
		} else {
			panic(fmt.Sprintf("Couldn't dial to lotus ws url despite %d retries. err: %s", maxDialReties, err))
		}
	}

	return conn
}

func (c *RPCClient) getNextId() uint64 {
	return atomic.AddUint64(&c.nextId, 1)
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
		Id:      c.getNextId(),
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
		_, msg, err := c.conn.ReadMessage()
		err = json.Unmarshal(msg, sub)
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

func (c *RPCClient) Do(method string, params []interface{}, dst interface{}) error {
	return c.do(method, params, dst, 0)
}

func (c *RPCClient) do(method string, params []interface{}, dst interface{}, retryCount int) error {
	request := &Request{
		Id:      c.getNextId(),
		Version: jsonRPCVersion,
		Method:  method,
		Params:  params,
	}

	err := c.conn.WriteJSON(request)
	if err != nil {
		log.Println("[RPCClient][Error][Subscribe]", err)
		return err
	}

	resChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	// Receiving first message after sending request, it should be type of dst.
	// Goroutine and select from channel are for timeout only.
	go func() {
		_, res, err := c.conn.ReadMessage()
		if err != nil {
			errChan <- err
		}
		resChan <- res
	}()

	select {
	case err := <-errChan:
		if retryCount < maxRequestRetries {
			retryCount++
			log.Println("[RPCClient][Debug][do]", "Received error on method", method, "err:", err, "Attempting retry number", retryCount)
			return c.do(method, params, dst, retryCount)
		}

		return fmt.Errorf("couldn't receive correct response on method %s after %d retries. err: %s", method, maxRequestRetries, err)

	case res := <-resChan:
		return json.Unmarshal(res, &dst)

	case <-time.After(wsRequestTimeout):
		if retryCount < maxRequestRetries {
			retryCount++
			log.Println("[RPCClient][Debug][do]", "Received timeout on method", method, "Attempting retry number", retryCount)
			return c.do(method, params, dst, retryCount)
		}

		return fmt.Errorf("request timeout on method %s after %d retries", method, maxRequestRetries)
	}
}
