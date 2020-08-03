package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/segmentio/kafka-go"
	"log"
)

const (
	kafkaPartition = 0
	topicBlocks    = "blocks_stream"
	topicMessages  = "messages_stream"
)

type Datastore struct {
	config       *config.Config
	kafkaWriters map[string]*kafka.Writer
	// ack   chan kafka.Event
	// pushChan chan interface{}
}

func Init(config *config.Config) (*Datastore, error) {
	ds := &Datastore{
		config:       config,
		kafkaWriters: make(map[string]*kafka.Writer),
		// pushChan:     make(chan interface{}),
	}

	for _, topic := range []string{topicBlocks, topicMessages} {
		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{ds.config.KafkaHosts},
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		})
		if writer == nil {
			return nil, errors.New("cannot create kafka writer")
		}
		ds.kafkaWriters[topic] = writer
	}

	return ds, nil
}

func (ds *Datastore) Push(key string, i interface{}) {
	ds.push(key, i)
}

func (ds *Datastore) push(key string, i interface{}) (err error) {
	var (
		topic string
	)
	// log.Println("[Datastore][push][Debug] Push data to kafka")

	if ds.kafkaWriters == nil {
		log.Println("[Datastore][Error][push]", "Kafka writers not initialized")
		return errors.New("cannot push")
	}

	switch i.(type) {
	case types.BlockHeader:
		topic = topicBlocks
	case types.Message:
		topic = topicMessages
	default:
		log.Println("[Datastore][Error][push]", "Unsupported struct")
		return errors.New("cannot push")
	}

	data, err := json.Marshal(i)
	if err != nil {
		log.Println("[Datastore][Error][push]", "Cannot marshal push data", err)
		return errors.New("cannot push")
	}

	if _, ok := ds.kafkaWriters[topic]; !ok {
		log.Println("[Datastore][Error][push]", "Kafka writer not initialized for topic", topic)
	}

	// TODO: Add key
	err = ds.kafkaWriters[topic].WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: data,
	})

	if err != nil {
		log.Println("[Datastore][Error][push]", "Cannot produce data", err)
	}
	return err
}
