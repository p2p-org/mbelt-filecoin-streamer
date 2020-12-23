package main

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"

	//"github.com/filecoin-project/specs-actors/actors/abi"
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func genAddress() address.Address {
	addr, err := address.NewActorAddress([]byte(randStringRunes(10)))
	if err != nil {
		return address.Address{}
	}
	return addr
}

func genMsg() *types.Message {
	return &types.Message{
		Version:  rand.Int63(),
		To:       genAddress(),
		From:     genAddress(),
		Nonce:    rand.Uint64(),
		Value:    types.NewInt(rand.Uint64()),
		GasPrice: types.NewInt(rand.Uint64()),
		GasLimit: rand.Int63(),
		Method:   abi.MethodNum(rand.Int63()),
		Params:   []byte(randStringRunes(10)),
	}
}

func main() {
	topic := "messages_stream"
	partition := 0

	log.Println("Connecting to kafka...")

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer conn.Close()
	log.Println("Connected!")

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	for {
		msg := genMsg()
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Fatalln(err.Error)
		}
		if _, err = conn.Write(msgBytes); err != nil {
			log.Fatalln(err.Error())
		}
		log.Println("Message sent")
	}
}
