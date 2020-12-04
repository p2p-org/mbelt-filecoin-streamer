package client

import (
	"log"
	"sync"
)

type WsClientPool struct {
	pool  *sync.Pool
}

func NewWsClientPool(wsUrl string) WsClientPool {
	return WsClientPool{
		pool: &sync.Pool{New: func() interface{} {
			client, err := NewClient(wsUrl)
			if err != nil {
				log.Println(err)
			}
			return client
		}},
	}
}

func (m *WsClientPool) Get() *RPCClient {
	return m.pool.Get().(*RPCClient)
}

func (m *WsClientPool) Put(client *RPCClient) {
	m.pool.Put(client)
}
